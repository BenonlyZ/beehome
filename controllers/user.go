package controllers

import (
	"ihome/models"
	"path"

	"github.com/astaxie/beego"
	"github.com/op-y/weilaihui/fdfs_client"

	"encoding/json"

	"github.com/astaxie/beego/orm"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) RetData(resp interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

// /api/v1.0/users [post]
func (c *UserController) Reg() {

	resp := make(map[string]interface{})

	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	defer c.RetData(resp)
	//1.得到客户端请求的json数据 post数据
	regRequestMap := make(map[string]interface{})
	json.Unmarshal(c.Ctx.Input.RequestBody, &regRequestMap)
	beego.Info("mobile = ", regRequestMap["mobile"], "passsword", regRequestMap["password"])
	//2.判断数据合法性
	if regRequestMap["mobile"] == "" || regRequestMap["password"] == "" || regRequestMap["sms_code"] == "" {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}
	//3.将数据存入到Mysql数据库 user表中
	user := models.User{}
	user.Mobile = regRequestMap["mobile"].(string)
	user.Password_hash = regRequestMap["password"].(string)
	user.Name = regRequestMap["mobile"].(string)

	o := orm.NewOrm()
	id, err := o.Insert(&user)
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		beego.Info("insert fail,id = ", id)
		return
	}
	c.SetSession("name", user.Mobile)
	c.SetSession("user_id", id)
	c.SetSession("mobile", user.Mobile)

}

func (this *UserController) Postavatar() {

	resp := make(map[string]interface{})
	defer this.RetData(resp)
	//1.获取上传的一个文件
	fileData, hd, err := this.GetFile("avatar")
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		beego.Info("===========11111")
		return
	}
	//2.得到文件后缀
	suffix := path.Ext(hd.Filename) //a.jpg.avi

	//3.存储文件到fastdfs上
	fdfsClient, err := fdfs_client.NewFdfsClient("conf/client.conf")
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		beego.Info("===========22222222")

		return
	}
	fileBuffer := make([]byte, hd.Size)
	_, err = fileData.Read(fileBuffer)
	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		beego.Info("===========33333")

		return
	}
	DataResponse, err := fdfsClient.UploadByBuffer(fileBuffer, suffix[1:]) //aa.jpg

	if err != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		beego.Info("===========44444")

		return
	}

	//4.从session里拿到user_id
	user_id := this.GetSession("user_id")
	var user models.User
	//5.更新用户数据库中的内容
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	qs.Filter("Id", user_id).One(&user)
	user.Avatar_url = DataResponse.RemoteFileId

	_, errUpdate := o.Update(&user)
	if errUpdate != nil {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}

	urlMap := make(map[string]string)
	//Avaurl := "127.0.0.1:8899"+DataResponse.RemoteFileId
	urlMap["avatar_url"] = "http://127.0.0.1:8899/" + DataResponse.RemoteFileId
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
	resp["data"] = urlMap

}

func (c *UserController) Login() {

	resp := make(map[string]interface{})

	defer c.RetData(resp)
	//1.得到客户端请求的json数据 post数据
	regRequestMap := make(map[string]interface{})
	json.Unmarshal(c.Ctx.Input.RequestBody, &regRequestMap)
	beego.Info("mobile = ", regRequestMap["mobile"], "passsword", regRequestMap["password"])
	//2.判断数据合法性
	if regRequestMap["mobile"] == "" || regRequestMap["password"] == "" {
		resp["errno"] = models.RECODE_REQERR
		resp["errmsg"] = models.RecodeText(models.RECODE_REQERR)
		return
	}
	//3.查询数据
	var user models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	qs.Filter("mobile", regRequestMap["mobile"]).One(&user)
	if user.Password_hash != regRequestMap["password"] {
		resp["errno"] = models.RECODE_NODATA
		resp["errmsg"] = models.RecodeText(models.RECODE_NODATA)
		return
	}
	//4.添加session
	c.SetSession("name", user.Mobile)
	c.SetSession("user_id", user.Id)
	c.SetSession("mobile", user.Mobile)
	//5.返回json数据给前端
	resp["errno"] = models.RECODE_OK
	resp["errmsg"] = models.RecodeText(models.RECODE_OK)
}
