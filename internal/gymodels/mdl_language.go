package gymodels

import (
	"gorm.io/gorm"
)

type Language struct {
	Id              uint   `gorm:"autoIncrement; primaryKey; unique; not null"`
	ExtName         string `gorm:"type:varchar(20); unique; not null; default:c"`
	DisplayName     string `gorm:"type:varchar(50); not null; default:unknown"`
	Enabled         bool   `gorm:"not null; default:1"`
	SyntaxName      string `gorm:"type:varchar(50); not null; default:c_cpp"`
	SourceName      string `gorm:"type:varchar(50); not null; default:appmain.c"`
	ExeName         string `gorm:"type:varchar(50); not null; default:appmain"`
	CompileCmd      string `gorm:"type:varchar(200); not null; default:gcc {{.WorkPath}}/{{.SourceName}}"`
	ExecCmd         string `gorm:"type:varchar(200); not null; default:{{.WorkPath}}/{{.SourceName}}"`
	EnableSandbox   bool   `gorm:"not null; default:1"`
	LimitMemory     bool   `gorm:"not null; default:0"`
	LimitSyscall    bool   `gorm:"not null; default:0"`
	PregReplaceFrom string `gorm:"type:text; not null; default:"`
	PregReplaceTo   string `gorm:"type:text; not null; default:"`
	ForbiddenKeys   string `gorm:"type:text; not null; default:"`
}

type LanguageModel struct {
	db *gorm.DB
}
