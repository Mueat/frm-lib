package db

import (
	"strings"

	"github.com/Mueat/frm-lib/util"
	"gorm.io/gorm"
)

type Model struct {
	ID        uint              `gorm:"primarykey" json:"id"`
	CreatedAt DataTypeDateStamp `json:"created_at,omitempty"`
	UpdatedAt DataTypeDateStamp `json:"updated_at,omitempty"`
}

type DeletedModel struct {
	Model
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// 分页数据结构
type Pagination struct {
	Total       int64       `json:"total"`        // 总条数
	CurrentPage int         `json:"current_page"` // 当前页码
	PerPage     int         `json:"per_page"`     // 每页条数
	NextPage    int         `json:"next_page"`    // 下一页
	PrePage     int         `json:"pre_page"`     // 上一页
	Data        interface{} `json:"data"`         // 数据
}

const (
	// 按照日期查询
	DATEQUERY_TYPE_DATE = "date"
	// 按照季度查询
	DATEQUERY_TYPE_QUARTER = "quarter"
)

// 按照日期查询
//
// 日期查询格式使用逗号分隔。
// 第一个参数为查询类型，date(按照日期查询)或则quarter(按照季度查询)。
// 如果第一个参数为date, 则第二个参数为起始日期，第三个参数为结束日期，日期的格式为 YYYY-MM-DD。
// 如果第二个参数为quarter，则第后面的参数分别代表要查询的季度，如 quarter,1,4 表示查询第一和第四季度。
func QueryDate(db *gorm.DB, filedName string, query string) *gorm.DB {
	qs := strings.Split(query, ",")
	if qs[0] == DATEQUERY_TYPE_DATE {
		if len(qs) > 1 && qs[1] != "" {
			if _, err := util.Strtotime("2006-01-02", qs[1]); err != nil {
				db = db.Where(filedName+" >= ?", qs[1])
			}
		}
		if len(qs) > 2 && qs[2] != "" {
			if ts, err := util.Strtotime("2006-01-02", qs[2]); err != nil {
				end := util.Date("2006-01-02", ts+3600*24)
				db = db.Where(filedName+" < ?", end)
			}
		}
	}
	if qs[0] == "quarter" {
		if len(qs) > 1 {
			quarters := make([]string, 0)
			quarters = append(quarters, qs[1:]...)
			db = db.Where("quarter("+filedName+") in (?)", quarters)
		}
	}
	return db
}

// 获取分页数据
// args第一个参数为当前页码，不传默认为1
// args第二个参数为每页条数，不传默认为20
func Paginate(db *gorm.DB, data interface{}, args ...int) (*Pagination, error) {
	var count int64 = 0
	err := db.Count(&count).Error
	if err != nil {
		return nil, err
	}
	var page int = 1
	var perPage int = 20
	if args[0] > 0 {
		page = args[0]
	}
	if len(args) > 1 && args[1] > 0 {
		perPage = args[1]
	}

	if int(count) > (page-1)*perPage {
		err = db.Offset((page - 1) * perPage).Limit(perPage).Find(data).Error
		if err != nil {
			return nil, err
		}
	}

	nextPage := 0
	if int(count) > page*perPage {
		nextPage = page + 1
	}
	prePage := 0
	if page > 1 {
		prePage = page - 1
	}

	rest := Pagination{
		Total:       count,
		CurrentPage: page,
		PerPage:     perPage,
		NextPage:    nextPage,
		PrePage:     prePage,
		Data:        data,
	}
	return &rest, nil
}
