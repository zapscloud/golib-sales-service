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

type RegionService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(regionId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(regionId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(regionId string, delete_permanent bool) error

	EndService()
}

type regionBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoRegion   sales_repository.RegionDao
	daoBusiness platform_repository.BusinessDao
	child       RegionService
	businessId  string
}

// NewRegionService - Construct Region
func NewRegionService(props utils.Map) (RegionService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("RegionService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := regionBaseService{}
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

// regionBaseService - Close all the services
func (p *regionBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *regionBaseService) initializeService() {
	log.Printf("RegionService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoRegion = sales_repository.NewRegionDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *regionBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("regionBaseService::FindAll - Begin")

	listdata, err := p.daoRegion.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("regionBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *regionBaseService) Get(regionId string) (utils.Map, error) {
	log.Printf("regionBaseService::Get::  Begin %v", regionId)

	data, err := p.daoRegion.Get(regionId)

	log.Println("regionBaseService::Get:: End ", err)
	return data, err
}

func (p *regionBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("RegionService::FindByCode::  Begin ", filter)

	data, err := p.daoRegion.Find(filter)
	log.Println("RegionService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *regionBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("RegionService::Create - Begin")
	var regionId string

	dataval, dataok := indata[sales_common.FLD_REGION_ID]
	if dataok {
		regionId = strings.ToLower(dataval.(string))
	} else {
		regionId = utils.GenerateUniqueId("rgn")
		log.Println("Unique Region ID", regionId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_REGION_ID] = regionId

	data, err := p.daoRegion.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("RegionService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *regionBaseService) Update(regionId string, indata utils.Map) (utils.Map, error) {

	log.Println("RegionService::Update - Begin")

	data, err := p.daoRegion.Update(regionId, indata)

	log.Println("RegionService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *regionBaseService) Delete(regionId string, delete_permanent bool) error {

	log.Println("RegionService::Delete - Begin", regionId)

	if delete_permanent {
		result, err := p.daoRegion.Delete(regionId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(regionId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("RegionService::Delete - End")
	return nil
}

func (p *regionBaseService) errorReturn(err error) (RegionService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
