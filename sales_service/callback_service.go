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

type CallbackService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(callbackId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(callbackId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(callbackId string, delete_permanent bool) error

	EndService()
}

type callbackBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoCallback sales_repository.CallbackDao
	daoBusiness platform_repository.BusinessDao
	child       CallbackService
	businessId  string
}

// NewCallbackService - Construct Callback
func NewCallbackService(props utils.Map) (CallbackService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CallbackService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := callbackBaseService{}
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

// callbackBaseService - Close all the service
func (p *callbackBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *callbackBaseService) initializeService() {
	log.Printf("CallbackService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCallback = sales_repository.NewCallbackDao(p.GetClient(), p.businessId)
}

// List - List All records
func (p *callbackBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("callbackBaseService::FindAll - Begin")

	listdata, err := p.daoCallback.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("callbackBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *callbackBaseService) Get(callbackId string) (utils.Map, error) {
	log.Printf("callbackBaseService::Get::  Begin %v", callbackId)

	data, err := p.daoCallback.Get(callbackId)

	log.Println("callbackBaseService::Get:: End ", err)
	return data, err
}

func (p *callbackBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("CallbackService::FindByCode::  Begin ", filter)

	data, err := p.daoCallback.Find(filter)
	log.Println("CallbackService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *callbackBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CallbackService::Create - Begin")
	var callbackId string

	dataval, dataok := indata[sales_common.FLD_CALLBACK_ID]
	if dataok {
		callbackId = strings.ToLower(dataval.(string))
	} else {
		callbackId = utils.GenerateUniqueId("clbk_")
		log.Println("Unique Callback ID", callbackId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CALLBACK_ID] = callbackId
	indata[sales_common.FLD_IS_FULFILLED] = false

	data, err := p.daoCallback.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CallbackService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *callbackBaseService) Update(callbackId string, indata utils.Map) (utils.Map, error) {

	log.Println("CallbackService::Update - Begin")

	data, err := p.daoCallback.Update(callbackId, indata)

	log.Println("CallbackService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *callbackBaseService) Delete(callbackId string, delete_permanent bool) error {

	log.Println("CallbackService::Delete - Begin", callbackId)

	indata := utils.Map{db_common.FLD_IS_DELETED: true}
	data, err := p.Update(callbackId, indata)
	if err != nil {
		return err
	}
	log.Println("Update for Delete Flag", data)

	log.Printf("CallbackService::Delete - End")
	return nil

}
func (p *callbackBaseService) errorReturn(err error) (CallbackService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
