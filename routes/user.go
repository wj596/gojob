/*
 * Copyright 2020-2021 the original author(https://github.com/wj596)
 *
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * </p>
 */
package routes

import (
	"net/http"
	"sync"

	"gojob/internal"
	"gojob/models"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/gin-gonic/gin"
)

// action of bolt

var tokens sync.Map

func insertUser(c *gin.Context) {
	user := new(models.User)
	err := c.BindJSON(user)
	if nil != err {
		respond400(c, err.Error())
		return
	}
	stock, _ := models.GetUser(user.Name)
	if nil != stock {
		respond400(c, "存在用户名为："+user.Name+" 的用户，请更换")
		return
	}
	err = internal.InsertUser(user)
	if nil != err {
		respond500(c, err.Error())
		return
	}
	respondOK(c)
}

func deleteUser(c *gin.Context) {
	id := stringutil.ToUintSafe(c.Param("id"))
	err := internal.DeleteUser(id)
	if nil != err {
		respond500(c, err.Error())
		return
	}
	respondOK(c)
}

func updateUser(c *gin.Context) {
	user := new(models.User)
	err := c.BindJSON(user)
	if nil != err {
		logs.Error(err.Error())
		respond400(c, err.Error())
		return
	}

	err = internal.UpdateUser(user)
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
		return
	}
	respondOK(c)
}

func getUser(c *gin.Context) {
	name := c.Param("name")
	user, err := models.GetUser(name)
	if nil != err {
		respond500(c, err.Error())
	} else {
		respondData(c, user)
	}
}

func searchUser(c *gin.Context) {
	if "" != c.Query("page_num") && "" != c.Query("page_size") {
		pageUser(c)
	} else {
		ps := models.NewCondition().AddParam("hasEmail", "true")
		respondData(c, models.SelectUserList(ps))
	}
}

func pageUser(c *gin.Context) {
	pageNum := c.Query("page_num")
	pageSize := c.Query("page_size")
	ps := models.NewCondition().AddParam("name", c.Query("name"))
	startIndex := (stringutil.ToIntSafe(pageNum) - 1) * stringutil.ToIntSafe(pageSize)
	list := models.SelectUserList(ps)
	slice := make([]*models.User, 0)
	for i := 0; i < stringutil.ToIntSafe(pageSize); i++ {
		index := startIndex + i
		if index < len(list) {
			vo := list[index]
			slice = append(slice, vo)
		}
	}
	respondPage(c, &models.Page{
		Total: int64(len(list)),
		Data:  slice,
	})
}

func login(c *gin.Context) {
	ps := new(models.User)
	err := c.BindJSON(ps)
	if nil != err {
		respond400(c, err.Error())
		return
	}
	cf, err := doLogin(ps.Name, ps.Password)
	if nil != err {
		respond400(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succeed":  true,
		"token":    cf.Token,
		"userName": cf.User.Name,
	})
}

func logout(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if "" != token {
		doLogout(token)
	}
	respondOK(c)
}

func authorised(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if "" == token {
		c.AbortWithStatus(401)
		return
	}
	cf, err := doAuthorised(token)
	if nil != err {
		c.AbortWithStatus(401)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succeed":  true,
		"token":    cf.Token,
		"userName": cf.User.Name,
	})
}
