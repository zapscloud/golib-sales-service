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

type FirmnessService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(firmnessId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(firmnessId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(firmnessId string, delete_permanent bool) error

	EndService()
}

type firmnessBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoFirmness sales_repository.FirmnessDao
	daoBusiness platform_repository.BusinessDao
	child       FirmnessService
	businessId  string
}

// NewFirmnessService - Construct Firmness
func NewFirmnessService(props utils.Map) (FirmnessService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("FirmnessService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := firmnessBaseService{}
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

// firmnesssBaseService - Close all the services
func (p *firmnessBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *firmnessBaseService) initializeService() {
	log.Printf("FirmnessService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoFirmness = sales_repository.NewFirmnessDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *firmnessBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("firmnessBaseService::FindAll - Begin")

	listdata, err := p.daoFirmness.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("firmnessBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *firmnessBaseService) Get(firmnessId string) (utils.Map, error) {
	log.Printf("firmnessBaseService::Get::  Begin %v", firmnessId)

	data, err := p.daoFirmness.Get(firmnessId)

	log.Println("firmnessBaseService::Get:: End ", err)
	return data, err
}

func (p *firmnessBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("FirmnessService::FindByCode::  Begin ", filter)

	data, err := p.daoFirmness.Find(filter)
	log.Println("FirmnessService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *firmnessBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("FirmnessService::Create - Begin")
	var firmnessId string

	dataval, dataok := indata[sales_common.FLD_FIRMNESS_ID]
	if dataok {
		firmnessId = strings.ToLower(dataval.(string))
	} else {
		firmnessId = utils.GenerateUniqueId("firms")
		log.Println("Unique Firmness ID", firmnessId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_FIRMNESS_ID] = firmnessId

	data, err := p.daoFirmness.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("FirmnessService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *firmnessBaseService) Update(firmnessId string, indata utils.Map) (utils.Map, error) {

	log.Println("FirmnessService::Update - Begin")

	data, err := p.daoFirmness.Update(firmnessId, indata)

	log.Println("FirmnessService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *firmnessBaseService) Delete(firmnessId string, delete_permanent bool) error {

	log.Println("FirmnessService::Delete - Begin", firmnessId)

	if delete_permanent {
		result, err := p.daoFirmness.Delete(firmnessId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(firmnessId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("FirmnessService::Delete - End")
	return nil
}

func (p *firmnessBaseService) errorReturn(err error) (FirmnessService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
