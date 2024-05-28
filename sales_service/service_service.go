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

// ServiceService - Business Service Service structure
type ServiceService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(serviceId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(serviceId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(serviceId string, delete_permanent bool) error

	EndService()
}

// ServiceService - Business Service Service structure
type serviceBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoService  sales_repository.ServiceDao
	daoBusiness platform_repository.BusinessDao
	child       ServiceService
	businessId  string
}

// NewServiceService - Construct Service
func NewServiceService(props utils.Map) (ServiceService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("ServiceService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := serviceBaseService{}
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
			ErrorMsg:    "Invalid business_id",
			ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// serviceBaseService - Close all the services
func (p *serviceBaseService) EndService() {
	log.Printf("EndServiceService ")
	p.CloseDatabaseService()
}

func (p *serviceBaseService) initializeService() {
	log.Printf("ServiceMongoService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoService = sales_repository.NewServiceDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *serviceBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("ServiceService::FindAll - Begin")

	listdata, err := p.daoService.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("ServiceService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *serviceBaseService) Get(serviceId string) (utils.Map, error) {
	log.Printf("ServiceService::Get::  Begin %v", serviceId)

	data, err := p.daoService.Get(serviceId)

	log.Println("ServiceService::Get:: End ", err)
	return data, err
}

func (p *serviceBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("serviceBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoService.Find(filter)
	log.Println("serviceBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *serviceBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("ServiceService::Create - Begin")
	var serviceId string

	dataval, dataok := indata[sales_common.FLD_SERVICE_ID]
	if dataok {
		serviceId = strings.ToLower(dataval.(string))
	} else {
		serviceId = utils.GenerateUniqueId("prod")
		log.Println("Unique Service ID", serviceId)
	}

	//BusinessService
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_SERVICE_ID] = serviceId

	data, err := p.daoService.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("ServiceService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *serviceBaseService) Update(serviceId string, indata utils.Map) (utils.Map, error) {

	log.Println("BusinessProdcutService::Update - Begin")

	data, err := p.daoService.Update(serviceId, indata)

	log.Println("ServiceService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *serviceBaseService) Delete(serviceId string, delete_permanent bool) error {

	log.Println("ServiceService::Delete - Begin", serviceId)

	if delete_permanent {
		result, err := p.daoService.Delete(serviceId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(serviceId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("ServiceService::Delete - End")
	return nil
}

func (p *serviceBaseService) errorReturn(err error) (ServiceService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
