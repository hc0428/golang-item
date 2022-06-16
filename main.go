package main

import (
	DR "ManageProject/DB"
	"ManageProject/conf"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"time"
)

var cfg = new(conf.AppConf)
var JwtKey = []byte("www.auth.com")

type Claims struct {
	Username string
	Power int
	jwt.StandardClaims
}

func SetToken(username string, power int) (string, error) {
	expireTime := time.Now().Add(5 * time.Hour)
	SetClaims := Claims{
		username,
		power,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer: "127.0.0.1",
			Subject: "user token",
		},
	}

	reqClaim := jwt.NewWithClaims(jwt.SigningMethodHS256, SetClaims)

	token, err := reqClaim.SignedString(JwtKey)
	if err != nil {
		return "", errors.New("token 生成失败")
	}

	return token, nil
}

func ParseToken(token string) (*Claims, error) {
	setToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil {
		errors.New("解析token失败")
	}

	if key, ok := setToken.Claims.(*Claims); ok && setToken.Valid {
		return key, nil
	} else {
		return nil, errors.New("获取相应token 信息失败")
	}
}

func MidJwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"message": "token not exist",
			})

			c.Abort()
			return
		}

		cla, err := ParseToken(token)

		if err != nil {
			fmt.Println(err)
			return
		}

		if cla.ExpiresAt < time.Now().Unix() {
			c.JSON(http.StatusOK, gin.H{
				"message": "登录过期",
			})
		}

		c.Set("power", cla.Power)
		c.Next()
	}
}

func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		var u DR.User

		db.Table("role").Where("name = ?", username).First(&u)

		if password == u.Password {
			if u.Role == "M" {
				token, err := SetToken(u.Name, 1)
				if err != nil {
					fmt.Println(err)
					c.JSON(http.StatusOK, gin.H{
						"errno": "110",
						"errmsg": "登录失败",
						"data": nil,
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "登录成功",
					"data": gin.H{
						"token": token,
						"role": "管理员",
					},
				})
			} else {
				token, err :=SetToken(u.Name, 0)

				if err != nil {
					fmt.Println(err)
					c.JSON(http.StatusOK, gin.H{
						"errno": "110",
						"errmsg": "登录失败",
						"data": nil,
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "登录成功",
					"data": gin.H{
						"token": token,
						"role": "普通用户",
					},
				})
			}

		}
	}
}

func main() {

	//读取相应的mysql配置并且进行连接

	err := ini.MapTo(cfg, "./conf/conf.ini")
	if err != nil {
		fmt.Println("ini mysql error: ", err)
		return
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?%s",
		cfg.Username,
		cfg.Password,
		"tcp",
		cfg.Host,
		cfg.Port,
		cfg.Database,
		"parseTime=true",
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		SkipDefaultTransaction: true,

	})


	if err != nil {
		fmt.Println("connect mysql error ", err)
		return
	}


	r := gin.Default()

	r.POST("/login", Login(db))

	other := r.Group("/auth", MidJwt())
	{
		other.GET("/book", func(c *gin.Context) {
			var books []DR.Book
			flag := 0

			name, ok := c.GetQuery("name")
			if !ok {
				flag = flag | 1
			}

			author, ok := c.GetQuery("author")
			if !ok {
				flag = flag | 2
			}

			switch flag {
			case 0:
				db.Table("book").Where("name = ? and author = ?", name, author).Find(&books)
			case 1:
				db.Table("book").Where("name = ?", name).Find(&books)
			case 2:
				db.Table("book").Where("author = ? ", author).Find(&books)
			case 3:
				db.Table("book").Find(&books)
			}

			data, err := json.Marshal(books)
			if err != nil {
				fmt.Println("json marshal fail", err)
				c.JSON(http.StatusOK, gin.H{
					"errno": "1010",
					"errmsg": "系统繁忙~请重试",
					"data": nil,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"errno": "0",
				"errmsg": "查询成功",
				"data": string(data),
			})
		})

		other.POST("/book", func(c *gin.Context) {
			power, ext := c.Get("power")
			if !ext {
				fmt.Println("获取相应权限失败")
				c.JSON(http.StatusOK, gin.H{
					"errno": "1010",
					"errmsg": "系统繁忙~请重试",
					"data": nil,
				})
				return
			}

			if power == 0 {
				c.JSON(http.StatusOK, gin.H{
					"errno": "2",
					"errmsg": "无相应权限",
					"data": nil,
				})
			} else {
				b, err := c.GetRawData()
				if err != nil {
					fmt.Println("插入书籍获取相应json数据时失败", err)
					c.JSON(http.StatusOK, gin.H{
						"errno": "3",
						"errmsg": "没输入相应查询词",
						"data": nil,
					})
					return
				}
				var book DR.Book

				err = json.Unmarshal(b, &book)
				if err != nil {
					fmt.Println("插入书籍 json unmarshal 失败", err)
					c.JSON(http.StatusOK, gin.H{
						"errno": "1010",
						"errmsg": "系统繁忙~请重试",
						"data": nil,
					})
					return
				}
				db.Table("book").Select("name", "author", "floor", "block", "bookshelf", "bookshelf_level", "exist").Create(&book)

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "SUCCESS",
					"data": nil,
				})
			}

		})

		other.PUT("/book", func(c *gin.Context) {
			power, ext := c.Get("power")
			if !ext {
				fmt.Println("获取相应权限失败")
				c.JSON(http.StatusOK, gin.H{
					"errno": "1010",
					"errmsg": "系统繁忙~请重试",
					"data": nil,
				})
				return
			}

			if power == 0 {
				c.JSON(http.StatusOK, gin.H{
					"errno": "2",
					"errmsg": "无相应权限",
					"data": nil,
				})
			} else {
				id, ok := c.GetQuery("id")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "12",
						"errmsg": "请输入id",
						"data": nil,
					})
				}

				ext, ok := c.GetQuery("exist")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "13",
						"errmsg": "请更新相应信息",
						"data": nil,
					})
				}

				db.Table("book").Where("id = ? ", id).Update("exist", ext)

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "SUCCESS",
					"data": nil,
				})
			}
		})

		other.PUT("/book/location", func(c *gin.Context) {
			power, ext := c.Get("power")
			if !ext {
				fmt.Println("获取相应权限失败")
				c.JSON(http.StatusOK, gin.H{
					"errno": "1010",
					"errmsg": "系统繁忙~请重试",
					"data": nil,
				})
				return
			}

			if power == 0 {
				c.JSON(http.StatusOK, gin.H{
					"errno": "2",
					"errmsg": "无相应权限",
					"data": nil,
				})
			} else {
				author, ok := c.GetQuery("author")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "4",
						"errmsg": "请输入作者",
						"data": nil,
					})
				}

				name, ok := c.GetQuery("name")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "5",
						"errmsg": "请输入书名",
						"data": nil,
					})
				}

				floor, ok := c.GetQuery("floor")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "6",
						"errmsg": "请输入楼层",
						"data": nil,
					})
				}

				block, ok := c.GetQuery("block")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "7",
						"errmsg": "请输入区域",
						"data": nil,
					})
				}

				bookshelf, ok := c.GetQuery("bookshelf")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "8",
						"errmsg": "请输入书架编号",
						"data": nil,
					})
				}

				bookshelf_level, ok := c.GetQuery("bookshelf_level")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "9",
						"errmsg": "请输入书架书层",
						"data": nil,
					})
				}


				db.Table("book").Where("name = ? and author = ? ", name, author).Updates(map[string]interface{}{"floor": floor, "block": block, "bookshelf": bookshelf, "bookshelf_level": bookshelf_level})

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "SUCCESS",
					"data": nil,
				})
			}
		})

		other.DELETE("/book", func(c *gin.Context) {
			power, ext := c.Get("power")
			if !ext {
				fmt.Println("获取相应权限失败")
				c.JSON(http.StatusOK, gin.H{
					"errno": "1010",
					"errmsg": "系统繁忙~请重试",
					"data": nil,
				})
				return
			}

			if power == 0 {
				c.JSON(http.StatusOK, gin.H{
					"errno": "2",
					"errmsg": "无相应权限",
					"data": nil,
				})
			} else {
				id, ok := c.GetQuery("id")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "10",
						"errmsg": "请输入相应书籍id进行删除",
						"data": nil,
					})
				}

				var book DR.Book

				db.Table("book").Delete(&book, id)

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "SUCCESS",
					"data": nil,
				})
			}
		})

		other.DELETE("/book/batch", func(c *gin.Context) {
			power, ext := c.Get("power")
			if !ext {
				fmt.Println("获取相应权限失败")
				c.JSON(http.StatusOK, gin.H{
					"errno": "1010",
					"errmsg": "系统繁忙~请重试",
					"data": nil,
				})
				return
			}

			if power == 0 {
				c.JSON(http.StatusOK, gin.H{
					"errno": "2",
					"errmsg": "无相应权限",
					"data": nil,
				})
			} else {
				name, ok := c.GetQuery("name")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "10",
						"errmsg": "请输入相应书籍名称进行删除",
						"data": nil,
					})
				}

				author, ok := c.GetQuery("author")
				if !ok {
					c.JSON(http.StatusOK, gin.H{
						"errno": "11",
						"errmsg": "请输入相应书籍作者进行删除",
						"data": nil,
					})
				}

				var book DR.Book

				db.Table("book").Where("name = ? and author = ?", name, author).Delete(&book)

				c.JSON(http.StatusOK, gin.H{
					"errno": "0",
					"errmsg": "SUCCESS",
					"data": nil,
				})
			}
		})
	}

	r.Run(":8080")

}