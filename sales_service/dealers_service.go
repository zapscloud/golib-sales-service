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

type DealerService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(dealerId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(dealerId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(dealerId string, delete_permanent bool) error

	EndService()
}

type dealerBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoDealer   sales_repository.DealerDao
	daoBusiness platform_repository.BusinessDao
	child       DealerService
	businessId  string
}

// NewDealerService - Construct Dealer
func NewDealerService(props utils.Map) (DealerService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("DealerService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := dealerBaseService{}
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

// dealerBaseService - Close all the services
func (p *dealerBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *dealerBaseService) initializeService() {
	log.Printf("DealerService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoDealer = sales_repository.NewDealerDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *dealerBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("dealerBaseService::FindAll - Begin")

	listdata, err := p.daoDealer.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("dealerBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *dealerBaseService) Get(dealerId string) (utils.Map, error) {
	log.Printf("dealerBaseService::Get::  Begin %v", dealerId)

	data, err := p.daoDealer.Get(dealerId)

	log.Println("dealerBaseService::Get:: End ", err)
	return data, err
}

func (p *dealerBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("dealerService::FindByCode::  Begin ", filter)

	data, err := p.daoDealer.Find(filter)
	log.Println("dealerService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *dealerBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("DealerService::Create - Begin")
	var dealerId string

	dataval, dataok := indata[sales_common.FLD_DEALER_ID]
	if dataok {
		dealerId = strings.ToLower(dataval.(string))
	} else {
		dealerId = utils.GenerateUniqueId("dealr")
		log.Println("Unique Dealer ID", dealerId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_DEALER_ID] = dealerId

	data, err := p.daoDealer.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("DealerService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *dealerBaseService) Update(dealerId string, indata utils.Map) (utils.Map, error) {

	log.Println("DealerService::Update - Begin")

	data, err := p.daoDealer.Update(dealerId, indata)

	log.Println("DealerService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *dealerBaseService) Delete(dealerId string, delete_permanent bool) error {

	log.Println("DealerService::Delete - Begin", dealerId)

	if delete_permanent {
		result, err := p.daoDealer.Delete(dealerId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(dealerId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("DealerService::Delete - End")
	return nil
}

func (p *dealerBaseService) errorReturn(err error) (DealerService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
