package service

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/iyacontrol/go-common/src/models"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db}
}

func (as *AuditService) Create(oa models.OperationAudit) error {
	oa.CreateTime = time.Now()
	return as.db.Create(&oa).Error
}

func (as *AuditService) List(page models.Page) (int64, []models.OperationAudit, error) {
	var oa []models.OperationAudit
	var count int64

	err := as.db.Model(&models.OperationAudit{}).Count(&count).
		Order("sgt_hawkeye_audit.create_time desc").
		Order("sgt_hawkeye_audit.id desc").
		//Select(`sgt_hawkeye_audit.*, sgt_hawkeye_group.name`).
		//Joins("LEFT JOIN sgt_hawkeye_group on sgt_hawkeye_group.id = sgt_hawkeye_audit.group_id").
		Limit(page.PageSize).
		Offset(page.Offset).Find(&oa).Error
	if err != nil {
		return count, oa, err
	}

	return count, oa, nil
}

func (as *AuditService) FindByShareID(sid string, page models.Page) (int64, []models.OperationAudit, error) {
	var oa []models.OperationAudit
	var count int64
	if result := as.db.Model(&models.OperationAudit{}).
		Order("sgt_hawkeye_audit.create_time desc").
		Where("share_id = ?", sid).
		Count(&count).
		Limit(page.PageSize).Offset(page.Offset).Find(&oa); result.Error != nil {
		return count, oa, errors.New(fmt.Sprintf("Can not find %v", sid))
	}

	return count, oa, nil
}
