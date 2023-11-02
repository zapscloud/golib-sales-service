package sales_services

import (
	"fmt"
	"log"
	"strings"

	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-sales-repository/sales_common"
	"github.com/zapscloud/golib-sales-repository/sales_repository"
	"github.com/zapscloud/golib-utils/utils"
)

type MediaService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(mediaId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(mediaId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(mediaId string, delete_permanent bool) error

	EndService()
}

type mediaBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoMedia    sales_repository.MediaDao
	daoBusiness platform_repository.BusinessDao
	child       MediaService
	businessId  string
}

// NewMediaService - Construct Media
func NewMediaService(props utils.Map) (MediaService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("MediaService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := mediaBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_services.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}

	// Assign the BusinessId
	p.businessId = businessId
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid BusinessId",
			ErrorDetail: "Given BusinessId is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// mediaBaseService - Close all the services
func (p *mediaBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *mediaBaseService) initializeService() {
	log.Printf("MediaService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoMedia = sales_repository.NewMediaDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *mediaBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("mediaBaseService::FindAll - Begin")

	listdata, err := p.daoMedia.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("mediaBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *mediaBaseService) Get(mediaId string) (utils.Map, error) {
	log.Printf("mediaBaseService::Get::  Begin %v", mediaId)

	data, err := p.daoMedia.Get(mediaId)

	log.Println("mediaBaseService::Get:: End ", err)
	return data, err
}

func (p *mediaBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("mediaBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoMedia.Find(filter)
	log.Println("mediaBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *mediaBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("MediaService::Create - Begin")
	var mediaId string

	dataval, dataok := indata[sales_common.FLD_MEDIA_ID]
	if dataok {
		mediaId = strings.ToLower(dataval.(string))
	} else {
		mediaId = utils.GenerateUniqueId("media")
		log.Println("Unique Media ID", mediaId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_MEDIA_ID] = mediaId

	data, err := p.daoMedia.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("MediaService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *mediaBaseService) Update(mediaId string, indata utils.Map) (utils.Map, error) {

	log.Println("MediaService::Update - Begin")

	data, err := p.daoMedia.Update(mediaId, indata)

	log.Println("MediaService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *mediaBaseService) Delete(mediaId string, delete_permanent bool) error {

	log.Println("MediaService::Delete - Begin", mediaId)

	if delete_permanent {
		result, err := p.daoMedia.Delete(mediaId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(mediaId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("MediaService::Delete - End")
	return nil
}

func (p *mediaBaseService) errorReturn(err error) (MediaService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
