package sandbox

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wuhan005/gadget"
	"github.com/wuhan005/govalid"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/db"
)

func ListSandboxesHandler(c *gin.Context) (int, interface{}) {
	sandboxes, err := db.Sandboxes.ListAll()
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to get sandboxes: %v", err)
	}
	return gadget.MakeSuccessJSON(sandboxes)
}

func GetSandboxHandler(c *gin.Context) (int, interface{}) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return gadget.MakeErrJSON(40000, "Input `id` must be a integer.")
	}

	sandbox, err := db.Sandboxes.GetByID(uint(id))
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to get sandbox: %v", err)
	}
	return gadget.MakeSuccessJSON(sandbox)
}

func CreateSandboxHandler(c *gin.Context) (int, interface{}) {
	var form struct {
		TemplateID  uint   `json:"template_id" valid:"required" label:"模板 ID"`
		Placeholder string `json:"placeholder"`
		Editable    bool   `json:"editable"`
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

	// Check template exists.
	_, err = db.Tpls.GetByID(form.TemplateID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gadget.MakeErrJSON(40400, "Template not found!")
		}
		return gadget.MakeErrJSON(50000, "Failed to get template data: %v", err)
	}

	err = db.Sandboxes.Create(db.CreateSandboxOptions{
		TemplateID:  0,
		Placeholder: "",
		Editable:    false,
	})
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to create sandbox: %v", err)
	}
	return gadget.MakeSuccessJSON("Create sandbox succeed!")
}

func UpdateTemplateHandler(c *gin.Context) (int, interface{}) {
	var form struct {
		ID          uint   `json:"id" valid:"required" label:"模板 ID"`
		TemplateID  uint   `json:"template_id" valid:"required" label:"模板 ID"`
		Placeholder string `json:"placeholder"`
		Editable    bool   `json:"editable"`
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

	// Check template exists.
	_, err = db.Tpls.GetByID(form.TemplateID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gadget.MakeErrJSON(40400, "Template not found!")
		}
		return gadget.MakeErrJSON(50000, "Failed to get template data: %v", err)
	}

	// Check sandbox exists.
	_, err = db.Sandboxes.GetByID(form.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gadget.MakeErrJSON(40400, "Sandbox not found!")
		}
		return gadget.MakeErrJSON(50000, "Failed to get sandbox data: %v", err)
	}

	err = db.Sandboxes.Update(db.UpdateSandboxOptions{
		ID:          form.ID,
		TemplateID:  form.TemplateID,
		Placeholder: form.Placeholder,
		Editable:    form.Editable,
	})
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to update sandbox: %v", err)
	}
	return gadget.MakeSuccessJSON("Update sandbox succeed!")
}

func DeleteTemplateHandler(c *gin.Context) (int, interface{}) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return gadget.MakeErrJSON(40000, "Input `id` must be a integer.")
	}

	_, err = db.Sandboxes.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gadget.MakeErrJSON(40400, "Sandbox not found!")
		}
		return gadget.MakeErrJSON(50000, "Failed to get sandbox data: %v", err)
	}

	err = db.Sandboxes.Delete(uint(id))
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to delete sandbox: %v", err)
	}
	return gadget.MakeSuccessJSON("Delete sandbox succeed.")
}
