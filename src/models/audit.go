package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ShortType = uint8

type GroupList []int

var _ sql.Scanner = &GroupList{}
var _ driver.Valuer = GroupList{}

//继承Scanner（Scan接受的是指针类型）
func (m *GroupList) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	b, _ := src.([]byte)
	return json.Unmarshal(b, m)
}

//继承Valuer（INSERT时，Valuer不接受指针类型）
func (m GroupList) Value() (driver.Value, error) {
	return json.Marshal(m)
}

const (
	OperatingTypeAdd ShortType = iota
	OperatingTypeUpdate
	OperatingTypeDelete
	OperatingTypeLogin
	OperatingTypeLogout
	OperatingTypeScale
	OperatingTypeRollback
	OperatingTypeBuild
	OperatingTypeRestart
)

const (
	OperatingResourceRule ShortType = iota
	OperatingResourceBusiness
	OperatingResourceInfra
	OperatingResourceReceiver
	OperatingResourceUser
	OperatingResourceGroup
	OperatingResourceRoute
	OperatingResourceProme
	OperatingResourceApp
	OperatingResourceSvc
	OperatingResourceCronJob
	OperatingResourceConfigMap
	OperatingResourceCluster
	OperatingResourceNamespace
	OperatingResourceHpa
	OperatingResourceNode
	OperatingResourceCanary
	OperatingResourceTemplate
	OperatingResourceApplication
	OperatingResourceStep
)

type OperationAudit struct {
	ID                int64     `json:"id"`
	ShareID           string    `json:"share_id"`
	GroupID           GroupList `json:"group_id"`
	OperatingObject   string    `json:"operating_object"`
	OperatingType     ShortType `json:"-"`
	OperatingResource ShortType `json:"-"`
	CreateTime        time.Time `json:"create_time"`

	Operation string `json:"operation" gorm:"-"`
	Resource  string `json:"resource" gorm:"-"`
	//Group     string `json:"group" gorm:"-"`
}

func (r *OperationAudit) TableName() string {
	return "sgt_hawkeye_audit"
}

func (o *OperationAudit) GetOperationResource() string {
	switch o.OperatingResource {
	case OperatingResourceRule:
		return "报警规则"
	case OperatingResourceBusiness:
		return "业务监控"
	case OperatingResourceInfra:
		return "基础监控"
	case OperatingResourceReceiver:
		return "报警接收"
	case OperatingResourceUser:
		return "用户"
	case OperatingResourceRoute:
		return "报警路由"
	case OperatingResourceGroup:
		return "用户组"
	case OperatingResourceProme:
		return "Prometheus 采集器"
	case OperatingResourceApp:
		return "应用"
	case OperatingResourceSvc:
		return "服务"
	case OperatingResourceCronJob:
		return "定时任务"
	case OperatingResourceConfigMap:
		return "配置"
	case OperatingResourceCluster:
		return "集群"
	case OperatingResourceNamespace:
		return "空间"
	case OperatingResourceHpa:
		return "自动扩缩"
	case OperatingResourceNode:
		return "节点"
	case OperatingResourceCanary:
		return "流量管理"
	case OperatingResourceTemplate:
		return "上线审批模板"
	case OperatingResourceApplication:
		return "上线审批申请"
	case OperatingResourceStep:
		return "上线审批流程"
	}
	return ""
}

func (o *OperationAudit) GetOperationType() string {
	switch o.OperatingType {
	case OperatingTypeAdd:
		return "添加"
	case OperatingTypeUpdate:
		return "更新"
	case OperatingTypeDelete:
		return "删除"
	case OperatingTypeLogin:
		return "登录"
	case OperatingTypeLogout:
		return "登出"
	case OperatingTypeScale:
		return "扩缩"
	case OperatingTypeRollback:
		return "回滚"
	case OperatingTypeBuild:
		return "构建"
	case OperatingTypeRestart:
		return "重启"
	}
	return ""
}

func OperationConversion(b string) (ShortType, ShortType, error) {
	bs := strings.Split(b, "-")
	ot, err := strconv.ParseUint(bs[0], 10, 8)
	if err != nil {
		return OperatingTypeAdd, OperatingResourceRule, err
	}

	or, err := strconv.ParseUint(bs[1], 10, 8)
	if err != nil {
		return OperatingTypeAdd, OperatingResourceRule, err
	}

	return ShortType(ot), ShortType(or), nil
}

func ConverOperationType(ot, or ShortType) string {
	return fmt.Sprintf("%d-%d", ot, or)
}
