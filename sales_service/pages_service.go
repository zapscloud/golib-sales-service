package sales_service

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

type PageService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(pageId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(pageId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(pageId string, delete_permanent bool) error

	EndService()
}

type pageBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoPage     sales_repository.PageDao
	daoBusiness platform_repository.BusinessDao
	child       PageService
	businessId  string
}

// NewPageService - Construct Page
func NewPageService(props utils.Map) (PageService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("PageService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := pageBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_service.OpenRegionDatabaseService(props)
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

// pagesBaseService - Close all the services
func (p *pageBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *pageBaseService) initializeService() {
	log.Printf("PageService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoPage = sales_repository.NewPageDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *pageBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("pageBaseService::FindAll - Begin")

	listdata, err := p.daoPage.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("pageBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *pageBaseService) Get(pageId string) (utils.Map, error) {
	log.Printf("pageBaseService::Get::  Begin %v", pageId)

	data, err := p.daoPage.Get(pageId)

	log.Println("pageBaseService::Get:: End ", err)
	return data, err
}

func (p *pageBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("pageBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoPage.Find(filter)
	log.Println("pageBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *pageBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("PageService::Create - Begin")
	var pageId string

	dataval, dataok := indata[sales_common.FLD_PAGE_ID]
	if dataok {
		pageId = strings.ToLower(dataval.(string))
	} else {
		pageId = utils.GenerateUniqueId("page")
		log.Println("Unique Page ID", pageId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_PAGE_ID] = pageId

	data, err := p.daoPage.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("PageService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *pageBaseService) Update(pageId string, indata utils.Map) (utils.Map, error) {

	log.Println("PageService::Update - Begin")

	data, err := p.daoPage.Update(pageId, indata)

	log.Println("PageService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *pageBaseService) Delete(pageId string, delete_permanent bool) error {

	log.Println("PageService::Delete - Begin", pageId)

	if delete_permanent {
		result, err := p.daoPage.Delete(pageId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(pageId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("PageService::Delete - End")
	return nil
}

func (p *pageBaseService) errorReturn(err error) (PageService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
