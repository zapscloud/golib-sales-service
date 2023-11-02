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

type StatesService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(statesId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(statesId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(statesId string, delete_permanent bool) error

	EndService()
}

type statesBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoStates   sales_repository.StatesDao
	daoBusiness platform_repository.BusinessDao
	child       StatesService
	businessId  string
}

// NewStatesService - Construct States
func NewStatesService(props utils.Map) (StatesService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("StatesService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := statesBaseService{}
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

// statesBaseService - Close all the services
func (p *statesBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *statesBaseService) initializeService() {
	log.Printf("StatesService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoStates = sales_repository.NewStatesDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *statesBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("statesBaseService::FindAll - Begin")

	listdata, err := p.daoStates.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("statesBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *statesBaseService) Get(statesId string) (utils.Map, error) {
	log.Printf("statesBaseService::Get::  Begin %v", statesId)

	data, err := p.daoStates.Get(statesId)

	log.Println("statesBaseService::Get:: End ", err)
	return data, err
}

func (p *statesBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("StatesService::FindByCode::  Begin ", filter)

	data, err := p.daoStates.Find(filter)
	log.Println("StatesService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *statesBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("StatesService::Create - Begin")
	var statesId string

	dataval, dataok := indata[sales_common.FLD_STATE_ID]
	if dataok {
		statesId = strings.ToLower(dataval.(string))
	} else {
		statesId = utils.GenerateUniqueId("stat")
		log.Println("Unique States ID", statesId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_STATE_ID] = statesId

	data, err := p.daoStates.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("StatesService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *statesBaseService) Update(statesId string, indata utils.Map) (utils.Map, error) {

	log.Println("StatesService::Update - Begin")

	data, err := p.daoStates.Update(statesId, indata)

	log.Println("StatesService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *statesBaseService) Delete(statesId string, delete_permanent bool) error {

	log.Println("StatesService::Delete - Begin", statesId)

	if delete_permanent {
		result, err := p.daoStates.Delete(statesId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(statesId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("StatesService::Delete - End")
	return nil
}

func (p *statesBaseService) errorReturn(err error) (StatesService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
