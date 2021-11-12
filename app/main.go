package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"
	"worstTwitter/bot"
	"worstTwitter/user"
	"worstTwitter/util"

	"github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("SESSION_USERID")
		if err != nil {
			return c.String(403, "please register")
		}
		id, err := util.UseSession(cookie.Value)
		if err != nil {
			return c.String(403, "please register")
		}

		user := user.FindUser(user.UserID(id))
		if user == nil {
			cookie.Expires.AddDate(1, 0, 0)
			c.SetCookie(cookie)
			return c.String(403, "invalid session")
		}
		c.Set("user", user)

		return next(c)
	}
}

func getIndex(c echo.Context) error {
	return c.Render(200, "index.html", "")
}

func getHome(c echo.Context) error {
	cuser := c.Get("user").(*user.User)
	return c.Render(200, "home.html", cuser)
}

func getUserName(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	user := user.FindUser(user.UserID(id))
	if user == nil {
		return c.NoContent(404)
	}
	return c.JSON(200, struct {
		Name string `json:"name"`
	}{Name: user.Name})
}

func getFollowRequests(c echo.Context) error {
	cuser := c.Get("user").(*user.User)
	return c.JSON(200, cuser.ReceivedFollowRequests)
}

func postAcceptFollowRequest(c echo.Context) error {
	cuser := c.Get("user").(*user.User)
	body := struct {
		FromID int `json:"from_id"`
	}{}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	err = user.FindUser(cuser.ID).AcceptFollowRequest(body.FromID)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func postTweet(c echo.Context) error {
	cuser := c.Get("user").(*user.User)

	body := struct {
		Text string `form:"tweet_text"`
	}{}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	cuser.Tweet(body.Text)
	return c.Redirect(303, "/home")
}

func postFollowRequest(c echo.Context) error {
	cuser := c.Get("user").(*user.User)
	body := struct {
		ToID int `json:"to_id"`
	}{}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	cuser.SendFollowRequest(user.UserID(body.ToID))
	return c.NoContent(200)
}

func postRegister(c echo.Context) error {
	username := c.FormValue("username")

	staffID := user.RegisterUser("staff_"+username, true, -1)
	id := user.RegisterUser(username, false, staffID)

	// staff is always watching you...
	user.FindUser(staffID).SendFollowRequest(id)
	user.FindUser(id).AcceptFollowRequest(staffID)

	c.SetCookie(&http.Cookie{
		Name:     "SESSION_USERID",
		Value:    util.Sign(id),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Expires:  time.Now().AddDate(0, 0, 1),
	})

	return c.Redirect(303, "/home")
}

type Template struct{}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	nonce, err := util.MakeRandomStr(10)
	if err != nil {
		return err
	}
	c.Response().Header().Add("Content-Security-Policy",
		fmt.Sprintf("default-src 'none'; frame-src www.google.com; connect-src 'self'; script-src 'nonce-%s'; style-src 'nonce-%s'; base-uri 'none'", nonce, nonce))

	temp, err := template.New(name).Funcs(template.FuncMap{
		"getnonce": func() string { return nonce },
	}).ParseFiles("templates/" + name)
	if err != nil {
		c.Logger().Errorf(err.Error())
		return err
	}
	return temp.Execute(w, data)
}

func postReport(c echo.Context) error {
	cuser := c.Get("user").(*user.User)

	bot.ReportChan <- cuser.StaffForThisUser
	return c.String(200, "please wait a second")
}

func main() {
	e := echo.New()
	t := &Template{}
	e.Renderer = t

	e.Use(middleware.Logger())

	e.GET("/", getIndex)
	e.GET("/users/:id/name", getUserName)

	e.POST("/register", postRegister)
	// TODO
	// e.POST("/login", postLogin)

	e.GET("/home", getHome, authMiddleware)

	e.POST("/tweet", postTweet, authMiddleware)
	e.POST("/report", postReport, authMiddleware)
	e.GET("/followreqs", getFollowRequests, authMiddleware)
	e.POST("/followreqs", postFollowRequest, authMiddleware)
	e.POST("/followreqs/accept", postAcceptFollowRequest, authMiddleware)

	bot.RunWorker()
	e.Logger.Fatal(e.Start(":8080"))
}
