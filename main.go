package main

import (
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/scrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Users 定义数据库的用户表
type Users struct {
	Username       string
	HashedPassword string
	Salt           string
}

// NewPostgres 数据库客户端
func NewPostgres() *gorm.DB {
	const dsn = "host=192.168.0.152 user=root password=msdnmm dbname=tests port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

func main() {
	http.HandleFunc("/signup", signup) // 注册
	http.HandleFunc("/login", login)   // 登录
	// 启动TCP服务
	if err := http.ListenAndServe(":4000", nil); err != nil {
		panic(err)
	}
}

// signup 注册
func signup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// 生成n个数的盐
	salt := generateSalt(16)
	// 对密码进行带盐的慢哈希值
	hashedPassword := hashPassword(password, salt)

	db := NewPostgres()

	user := &Users{
		Username:       username,
		HashedPassword: hashedPassword,
		Salt:           salt,
	}

	// 检查用户名是否被注册
	// 如果被注册, 返回已注册响应
	// 如果插入失败, 抛出错误
	// 反之返回注册成功
	if result := db.Table("users").FirstOrCreate(&user); result != nil {
		if result.RowsAffected == 0 {
			w.Write([]byte("注册失败, 该用户已被注册"))
		}
		w.Write([]byte("注册失败:" + result.Error.Error()))
	}
	w.Write([]byte("注册成功"))
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	db := NewPostgres()

	// 根据用户名获取该用户名的带盐慢哈希的密码值和盐值
	// 如果查不到用户或者查询异常, 抛出错误
	store, err := getHashedPasswordForUser(db, username)
	if err != nil {
		log.Printf("错误的查询, 错误消息: %v", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized) // 报401 未授权的访问
		return
	}

	// 将客户端传递的密码进行hash得到带盐慢哈希值
	verify := hashPassword(password, store.Salt)
	// 与数据库的带盐慢哈希的密码值进行比对
	if store.HashedPassword != verify {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized) // 报401 未授权的访问
		return
	}
	w.Write([]byte("OK"))
}

// 生成随机字符串作为盐值
func generateSalt(n int) string {
	// 初始化 Go 的伪随机数生成器
	source := rand.NewSource(time.Now().UnixNano())
	_ = rand.New(source)

	// 定义了一个包含小写字母和大写字母的 Unicode 符文rune切片, 作为随机构建字符串
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	// 创建n个长度的Unicode的字符串切片
	b := make([]rune, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 对密码进行bcrypt哈希并返回哈希值与随机盐值
func hashPassword(password, salt string) string {
	dk, err := scrypt.Key([]byte(password), []byte(salt), 1<<15, 8, 1, 32)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(dk)
}

// 从数据库获取用户哈希后的密码和盐值
// 如果数据库不存在用户名或者查询时的其他错误时抛出错误
func getHashedPasswordForUser(db *gorm.DB, username string) (*Users, error) {
	user := &Users{}
	if row := db.Table("users").Where("username = ?", username).First(user).Error; row != nil {
		if errors.Is(row, gorm.ErrRecordNotFound) {
			return nil, errors.New("数据库不存在此账号")
		}
		return nil, errors.New("查询失败, 错误消息:" + row.Error())
	}
	return user, nil
}
