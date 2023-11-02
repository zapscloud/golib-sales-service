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

type NavigationService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(navigationId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(navigationId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(navigationId string, delete_permanent bool) error

	EndService()
}

type navigationBaseService struct {
	db_utils.DatabaseService
	dbRegion      db_utils.DatabaseService
	daoNavigation sales_repository.NavigationDao
	daoBusiness   platform_repository.BusinessDao
	child         NavigationService
	businessId    string
}

// NewNavigationService - Construct Navigation
func NewNavigationService(props utils.Map) (NavigationService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("NavigationService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := navigationBaseService{}
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

// navigationBaseService - Close all the services
func (p *navigationBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *navigationBaseService) initializeService() {
	log.Printf("NavigationService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoNavigation = sales_repository.NewNavigationDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *navigationBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("navigationBaseService::FindAll - Begin")

	listdata, err := p.daoNavigation.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("navigationBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *navigationBaseService) Get(navigationId string) (utils.Map, error) {
	log.Printf("navigationBaseService::Get::  Begin %v", navigationId)

	data, err := p.daoNavigation.Get(navigationId)

	log.Println("navigationBaseService::Get:: End ", err)
	return data, err
}

func (p *navigationBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("NavigationService::FindByCode::  Begin ", filter)

	data, err := p.daoNavigation.Find(filter)
	log.Println("NavigationService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *navigationBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("NavigationService::Create - Begin")
	var navigationId string

	dataval, dataok := indata[sales_common.FLD_NAVIGATION_ID]
	if dataok {
		navigationId = strings.ToLower(dataval.(string))
	} else {
		navigationId = utils.GenerateUniqueId("nav")
		log.Println("Unique Navigation ID", navigationId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_NAVIGATION_ID] = navigationId

	data, err := p.daoNavigation.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("NavigationService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *navigationBaseService) Update(navigationId string, indata utils.Map) (utils.Map, error) {

	log.Println("NavigationService::Update - Begin")

	data, err := p.daoNavigation.Update(navigationId, indata)

	log.Println("NavigationService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *navigationBaseService) Delete(navigationId string, delete_permanent bool) error {

	log.Println("NavigationService::Delete - Begin", navigationId)

	if delete_permanent {
		result, err := p.daoNavigation.Delete(navigationId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(navigationId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("NavigationService::Delete - End")
	return nil
}

func (p *navigationBaseService) errorReturn(err error) (NavigationService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
