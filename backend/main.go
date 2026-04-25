package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ErrInfo struct {
	Code    int    `json:"code"`
	English string `json:"english"`
	Module  string `json:"module"`
	Message string `json:"message"`
}

type Err struct {
	Code   string
	Module string
	Msg    string
}

var (
	E50000 = Err{"E50000", "common", "服务器内部错误"}
	E40001 = Err{"E40001", "login", "用户名或密码错误"}
	E40002 = Err{"E40002", "register", "用户已存在"}
	E40003 = Err{"E40003", "account", "账户已存在"}
	E40004 = Err{"E40004", "account", "记录不存在"}
	E401   = Err{"401", "common", "未登录或登录失效"}
	E400   = Err{"400", "common", "参数错误"}
)

type BaseModel struct {
	Id        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;->;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updatedAt"`
}

type UserInfo struct {
	BaseModel
	LoginName string `gorm:"type:varchar(128);not null;unique;comment:登录名" json:"loginName"`
	Dek       string `gorm:"type:varchar(128);comment:数据密码(bcrypt)" json:"dek"`
}

type AuthRecord struct {
	BaseModel
	AccountId   uint      `gorm:"not null;index" json:"accountId"`
	Key         string    `gorm:"not null;index" json:"key"`
	BeforeValue string    `gorm:"type:varchar(255)" json:"beforeValue"`
	Value       string    `gorm:"not null;type:varchar(255)" json:"value"`
	IsEncrypt   bool      `gorm:"default:false" json:"isEncrypt"`
	IsSystem    bool      `gorm:"default:false" json:"isSystem"`
	IsTotp      bool      `gorm:"default:false" json:"isTotp"`
	IsDeleted   bool      `gorm:"default:false" json:"isDeleted"`
	Time        time.Time `gorm:"type:datetime;index" json:"time"`
}

type AuthAccount struct {
	BaseModel
	Code        string       `gorm:"type:varchar(6);uniqueIndex;not null" json:"code"`
	AccountName string       `gorm:"type:varchar(24);not null" json:"accountName"`
	Category    string       `json:"category"`
	AuthRecords []AuthRecord `gorm:"foreignKey:AccountId" json:"authRecords"`
}

var db *gorm.DB

func initDB() {
	var err error

	// 从环境变量读取数据库配置
	dbUser := os.Getenv("AUTHNOTE_DB_USER")
	dbPassword := os.Getenv("AUTHNOTE_DB_PASSWORD")
	dbHost := os.Getenv("AUTHNOTE_DB_HOST")
	dbPort := os.Getenv("AUTHNOTE_DB_PORT")
	dbName := os.Getenv("AUTHNOTE_DB_NAME")

	// MySQL 连接字符串 格式：用户名:密码@tcp(IP:端口)/数据库名?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Asia%%2FShanghai",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("MySQL数据库连接失败：" + err.Error())
	}

	err = db.AutoMigrate(&AuthAccount{}, &UserInfo{}, &AuthRecord{})

	if err != nil {
		panic("MySQL数据库自动建表失败：" + err.Error())
	}
}

func SuccessResult(c *gin.Context, data gin.H) {
	c.JSON(200, gin.H{
		"code": "200",
		"msg":  "success",
		"data": data,
	})
}

func ErrResult(c *gin.Context, err Err) {
	c.JSON(200, gin.H{
		"code": err.Code,
		"msg":  err.Msg,
		"data": nil,
	})
}

func Err400Result(c *gin.Context) {
	c.JSON(400, gin.H{
		"code": E400.Code,
		"msg":  E400.Msg,
		"data": nil,
	})
}

func Err401Result(c *gin.Context) {
	c.JSON(401, gin.H{
		"code": E401.Code,
		"msg":  E401.Msg,
		"data": nil,
	})
}

func ErrorResponse(c *gin.Context, msg string, code ...string) {
	resCode := "50000"
	if len(code) > 0 {
		resCode = code[0]
	}
	c.JSON(200, gin.H{
		"code": resCode,
		"msg":  msg,
		"data": nil,
	})
}

func AddAuthRecord(db *gorm.DB, accountId uint, key string, beforeValue string, value string, isEncrypt bool, time time.Time, isSystem bool) error {
	record := AuthRecord{
		AccountId:   accountId,
		Key:         key,
		BeforeValue: beforeValue,
		Value:       value,
		IsEncrypt:   isEncrypt,
		Time:        time,
		IsSystem:    isSystem,
	}
	return db.Create(&record).Error
}

type LoginRequest struct {
	LoginName string `json:"loginName"`
	Dek       string `json:"dek"`
}

type UserController struct{}

func IsBindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		Err400Result(c)
		return false
	}
	return true
}

// 检查用户名是否已存在
// true：用户名已存在，false：用户名不存在
func IsLoginNameExists(db *gorm.DB, loginName string) bool {
	var count int64
	db.Model(&UserInfo{}).Where("login_name = ?", loginName).Count(&count)
	return count > 0
}

func CheckUserDek(db *gorm.DB, loginName string, dek string) bool {
	var count int64
	db.Model(&UserInfo{}).
		Where("login_name = ? AND dek = ?", loginName, dek).
		Count(&count)
	return count > 0
}

func (u *UserController) UserLogin(c *gin.Context) {
	var req LoginRequest
	if !IsBindJSON(c, &req) {
		return
	}
	if !IsLoginNameExists(db, req.LoginName) {
		ErrResult(c, E40001)
		return
	}
	if !CheckUserDek(db, req.LoginName, req.Dek) {
		ErrResult(c, E40001)
		return
	}
	SuccessResult(c, gin.H{"login": true})
}

func (u *UserController) UserRegister(c *gin.Context) {
	var req LoginRequest
	if !IsBindJSON(c, &req) {
		return
	}
	if IsLoginNameExists(db, req.LoginName) {
		ErrResult(c, E40002)
		return
	}
	user := UserInfo{
		LoginName: req.LoginName,
		Dek:       req.Dek,
	}
	if err := db.Create(&user).Error; err != nil {
		ErrResult(c, E50000)
		return
	}
	SuccessResult(c, gin.H{"register": true})
}

type AccountController struct{}

func (ac *AccountController) Create(c *gin.Context) {
	ac.accountEdit(c, 0) // 0 = 新增
}
func (ac *AccountController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Err400Result(c)
		return
	}
	ac.accountEdit(c, uint(id)) // 传入ID = 修改
}
func (ac *AccountController) accountEdit(c *gin.Context, id uint) {
	var req AuthAccount
	if err := c.ShouldBindJSON(&req); err != nil {
		Err400Result(c)
		return
	}

	if req.Code == "" {
		Err400Result(c)
		return
	}

	// 检查重复
	query := db.Model(&AuthAccount{}).Where("code = ?", req.Code)
	if id > 0 {
		query = query.Where("id != ?", id)
	}

	var count int64
	query.Count(&count)
	if count > 0 {
		ErrResult(c, E40003)
		return
	}

	if id == 0 {
		// 新增
		if err := db.Create(&req).Error; err != nil {
			ErrorResponse(c, "创建失败")
			return
		}
		AddAuthRecord(db, req.Id, "创建账号", "", req.AccountName, false, time.Now(), true)
	} else {
		// 修改
		var oldAccount AuthAccount
		if err := db.First(&oldAccount, id).Error; err != nil {
			ErrResult(c, E40004)
			return
		}
		req.Id = id
		if err := db.Save(&req).Error; err != nil {
			ErrorResponse(c, "修改失败")
			return
		}
		AddAuthRecord(db, req.Id, "修改账号", oldAccount.AccountName, req.AccountName, false, time.Now(), true)
	}

	SuccessResult(c, gin.H{"edit": "success"})
}

type AuthRecordController struct {
	db *gorm.DB
}

func NewAuthRecordController(db *gorm.DB) *AuthRecordController {
	return &AuthRecordController{db: db}
}

func (arc *AuthRecordController) Create(c *gin.Context) {
	arc.recordEdit(c, 0, 1)
}

func (arc *AuthRecordController) Update(c *gin.Context) {
	itemIdStr := c.Param("itemid")
	itemId, err := strconv.Atoi(itemIdStr)
	if err != nil {
		Err400Result(c)
		return
	}
	arc.recordEdit(c, uint(itemId), 2)
}

func (arc *AuthRecordController) CreateUpdate(c *gin.Context) {
	itemIdStr := c.Param("itemid")
	itemId, err := strconv.Atoi(itemIdStr)
	if err != nil {
		Err400Result(c)
		return
	}
	arc.recordEdit(c, uint(itemId), 3)
}

func (arc *AuthRecordController) recordEdit(c *gin.Context, id uint, mode uint) {
	var req AuthRecord
	// 绑定前端参数
	if err := c.ShouldBindJSON(&req); err != nil {
		Err400Result(c)
		return
	}

	if req.AccountId == 0 {
		Err400Result(c)
		return
	}
	if req.Key == "" {
		Err400Result(c)
		return
	}
	if req.Value == "" {
		Err400Result(c)
		return
	}

	if id == 0 {
		// 检查同一个账户下是否已存在相同键名
		var count int64
		arc.db.Model(&AuthRecord{}).
			Where("account_id = ?", req.AccountId).
			Where("key = ? AND is_deleted = ?", req.Key, false).
			Count(&count)

		if count > 0 {
			ErrResult(c, E40003)
			return
		}

		// 创建
		if err := arc.db.Create(&req).Error; err != nil {
			ErrorResponse(c, "创建失败")
			return
		}

		SuccessResult(c, gin.H{"id": req.Id, "msg": "success"})
		return
	}

	var oldRecord AuthRecord
	if err := arc.db.First(&oldRecord, id).Error; err != nil {
		ErrResult(c, E40004)
		return
	}

	// 保存旧值
	req.BeforeValue = oldRecord.Value
	if mode == 3 {
		// 创建
		if err := arc.db.Create(&req).Error; err != nil {
			ErrorResponse(c, "修改失败")
			return
		}
	} else {
		req.Id = id
		// 更新
		if err := arc.db.Save(&req).Error; err != nil {
			ErrorResponse(c, "修改失败")
			return
		}
	}

	SuccessResult(c, gin.H{"msg": "success"})
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Origin, Content-Type, Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/api/test", func(c *gin.Context) {
		SuccessResult(c, gin.H{})
	})
	userController := &UserController{}
	api := r.Group("/api")
	api.POST("/register", userController.UserRegister)
	api.POST("/login", userController.UserLogin)
	accountController := AccountController{}
	r.POST("/api/accounts", accountController.Create)
	r.PUT("/api/accounts/:id", accountController.Update)
	recordCtrl := NewAuthRecordController(db)
	r.POST("/api/accounts/records", recordCtrl.Create)
	r.POST("/api/accounts/records/:itemid", recordCtrl.CreateUpdate)
	r.POST("/api/accounts/records/addlog/:id", func(c *gin.Context) {
		var req AuthRecord
		if err := c.ShouldBindJSON(&req); err != nil {
			Err400Result(c)
			return
		}
		AddAuthRecord(db, req.AccountId, "检查数据", "", req.Key, false, time.Now(), true)
		SuccessResult(c, gin.H{})
	})
	r.GET("/api/accounts", func(c *gin.Context) {
		// 1. 分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 10000 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize

		// 2. 子查询：每个账户最新一条记录
		subQuery := db.Model(&AuthRecord{}).
			Select("MAX(id) as id").
			Where("is_system = ? AND is_deleted = ?", false, false).
			Group("account_id, `key`")

		// 3. 查询账户 + 预加载最新记录
		var accounts []AuthAccount
		query := db.
			Preload("AuthRecords", "id IN (?)", subQuery).
			Order("id DESC").
			Limit(pageSize).
			Offset(offset)

		if err := query.Find(&accounts).Error; err != nil {
			ErrorResponse(c, "查询失败")
			return
		}

		// 5. 总数
		var total int64
		db.Model(&AuthAccount{}).Count(&total)

		// 6. 返回
		SuccessResult(c, gin.H{
			"list":       accounts,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		})
	})
	r.GET("/api/accounts/records", func(c *gin.Context) {
		var list []AuthRecord
		db.Find(&list)
		c.JSON(http.StatusOK, gin.H{"records": list})
	})

	r.GET("/api/accounts/:id/records", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var acc AuthAccount
		db.First(&acc, id)

		var list []AuthRecord
		db.Where("code = ?", acc.Code).Find(&list)
		c.JSON(http.StatusOK, gin.H{"records": list})
	})

	r.GET("/api/accounts/:id/records/:itemid", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		itemId, _ := strconv.Atoi(c.Param("itemid"))

		var acc AuthAccount
		db.First(&acc, id)

		var list []AuthRecord
		db.Where("code = ? AND id = ?", acc.Code, itemId).Find(&list)
		c.JSON(http.StatusOK, gin.H{"records": list})
	})

	_ = r.Run(":" + os.Getenv("AUTHNOTE_PORT"))
}
