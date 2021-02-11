package template

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuhan005/gadget"
	"github.com/wuhan005/govalid"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/db"
)

func ListTemplatesHandler(c *gin.Context) (int, interface{}) {
	templates, err := db.Tpls.ListAll()
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to get templates: %v", err)
	}
	return gadget.MakeSuccessJSON(templates)
}

func GetTemplateHandler(c *gin.Context) (int, interface{}) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return gadget.MakeErrJSON(40000, "Input `id` must be a integer.")
	}

	template, err := db.Tpls.GetByID(uint(id))
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to get template: %v", err)
	}
	return gadget.MakeSuccessJSON(template)
}

func CreateTemplateHandler(c *gin.Context) (int, interface{}) {
	var form struct {
		Name              string            `json:"name" valid:"required" label:"模板名称"`
		Language          []string          `json:"language" valid:"required" label:"编程语言"`
		Timeout           int               `json:"timeout" valid:"required;min:0;max:60" label:"超时时间"`
		MaxCPUs           int64             `json:"max_cpus" valid:"required;min:0;max:10" label:"最大 CPU 数"`
		MaxMemory         int64             `json:"max_memory" valid:"required;min:6;max:2048" label:"最大内存"` // MB
		InternetAccess    bool              `json:"internet_access"`
		DNS               map[string]string `json:"dns"`
		MaxContainer      int               `json:"max_container" valid:"required;min:0;max:1000" label:"最大容器数"`
		MaxContainerPerIP int               `json:"max_container_per_ip" valid:"required;min:0;max:100" label:"单 IP 最大容器数"`
	}
	err := c.BindJSON(&form)
	if err != nil {
		return gadget.MakeErrJSON(40300, "Failed to parse input JSON: %v", err)
	}

	v := govalid.New(form)
	if !v.Check() {
		for _, err := range v.Errors {
			return gadget.MakeErrJSON(40301, err.Message)
		}
	}

	err = db.Tpls.Create(db.CreateTplOptions{
		Name:              form.Name,
		Language:          form.Language,
		Timeout:           form.Timeout,
		MaxCPUs:           form.MaxCPUs,
		MaxMemory:         form.MaxMemory,
		InternetAccess:    form.InternetAccess,
		DNS:               form.DNS,
		MaxContainer:      form.MaxContainer,
		MaxContainerPerIP: form.MaxContainerPerIP,
	})
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to create template: %v", err)
	}
	return gadget.MakeSuccessJSON("Create template succeed!")
}

func UpdateTemplateHandler(c *gin.Context) (int, interface{}) {
	var form struct {
		ID                uint              `json:"id" valid:"required" label:"模板 ID"`
		Name              string            `json:"name" valid:"required" label:"模板名称"`
		Language          []string          `json:"language" valid:"required" label:"编程语言"`
		Timeout           int               `json:"timeout" valid:"required;min:0;max:60" label:"超时时间"`
		MaxCPUs           int64             `json:"max_cpus" valid:"required;min:0;max:10" label:"最大 CPU 数"`
		MaxMemory         int64             `json:"max_memory" valid:"required;min:6;max:2048" label:"最大内存"` // MB
		InternetAccess    bool              `json:"internet_access"`
		DNS               map[string]string `json:"dns"`
		MaxContainer      int               `json:"max_container" valid:"required;min:0;max:1000" label:"最大容器数"`
		MaxContainerPerIP int               `json:"max_container_per_ip" valid:"required;min:0;max:100" label:"单 IP 最大容器数"`
	}
	err := c.BindJSON(&form)
	if err != nil {
		return gadget.MakeErrJSON(40300, "Failed to parse input JSON: %v", err)
	}

	v := govalid.New(form)
	if !v.Check() {
		for _, err := range v.Errors {
			return gadget.MakeErrJSON(40301, err.Message)
		}
	}

	_, err = db.Tpls.GetByID(form.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gadget.MakeErrJSON(40400, "Template not found!")
		}
		return gadget.MakeErrJSON(50000, "Failed to get template data: %v", err)
	}

	err = db.Tpls.Update(db.UpdateTplOptions{
		ID:                form.ID,
		Name:              form.Name,
		Language:          form.Language,
		Timeout:           form.Timeout,
		MaxCPUs:           form.MaxCPUs,
		MaxMemory:         form.MaxMemory,
		InternetAccess:    form.InternetAccess,
		DNS:               form.DNS,
		MaxContainer:      form.MaxContainer,
		MaxContainerPerIP: form.MaxContainerPerIP,
	})
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to update template: %v", err)
	}
	return gadget.MakeSuccessJSON("Update template succeed!")
}

func DeleteTemplateHandler(c *gin.Context) (int, interface{}) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return gadget.MakeErrJSON(40000, "Input `id` must be a integer.")
	}

	_, err = db.Tpls.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gadget.MakeErrJSON(40400, "Template not found!")
		}
		return gadget.MakeErrJSON(50000, "Failed to get template data: %v", err)
	}

	err = db.Tpls.Delete(uint(id))
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to delete template: %v", err)
	}
	return gadget.MakeSuccessJSON("Delete template succeed.")
}
