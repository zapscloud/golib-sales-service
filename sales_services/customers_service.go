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

type CustomerService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(customerId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(customerId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(customerId string, delete_permanent bool) error

	// Authenticate Customer
	Authenticate(auth_key string, auth_login string, auth_pwd string) (utils.Map, error)
	// Change Password
	ChangePassword(userid string, newpwd string) (utils.Map, error)

	EndService()
}

type customerBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoCustomer sales_repository.CustomerDao
	daoBusiness platform_repository.BusinessDao
	child       CustomerService
	businessId  string
}

// NewCustomerService - Construct Customer
func NewCustomerService(props utils.Map) (CustomerService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CustomerService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := customerBaseService{}
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

// EndService - Close all the services
func (p *customerBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *customerBaseService) initializeService() {
	log.Printf("CustomerService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCustomer = sales_repository.NewCustomerDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *customerBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("customerBaseService::FindAll - Begin")

	listdata, err := p.daoCustomer.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("customerBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *customerBaseService) Get(customerId string) (utils.Map, error) {
	log.Printf("customerBaseService::Get::  Begin %v", customerId)

	data, err := p.daoCustomer.Get(customerId)

	// Delete the Password
	delete(data, sales_common.FLD_CUSTOMER_PASSWORD)

	log.Println("customerBaseService::Get:: End ", err)
	return data, err
}

func (p *customerBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("CustomerService::FindByCode::  Begin ", filter)

	data, err := p.daoCustomer.Find(filter)
	log.Println("CustomerService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *customerBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CustomerService::Create - Begin")
	var customerId string

	dataval, dataok := indata[sales_common.FLD_CUSTOMER_ID]
	if dataok {
		customerId = strings.ToLower(dataval.(string))
	} else {
		customerId = utils.GenerateUniqueId("cust")
		log.Println("Unique Customer ID", customerId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CUSTOMER_ID] = customerId

	// Hash the password if passed
	if dataVal, dataOk := indata[sales_common.FLD_CUSTOMER_PASSWORD]; dataOk {
		indata[sales_common.FLD_CUSTOMER_PASSWORD] = utils.SHA(dataVal.(string))
	}

	data, err := p.daoCustomer.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CustomerService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *customerBaseService) Update(customerId string, indata utils.Map) (utils.Map, error) {

	log.Println("CustomerService::Update - Begin")

	// Delete the Key fields if exist
	delete(indata, sales_common.FLD_BUSINESS_ID)
	delete(indata, sales_common.FLD_CUSTOMER_ID)

	// Hash the password if passed
	if dataVal, dataOk := indata[sales_common.FLD_CUSTOMER_PASSWORD]; dataOk {
		indata[sales_common.FLD_CUSTOMER_PASSWORD] = utils.SHA(dataVal.(string))
	}

	data, err := p.daoCustomer.Update(customerId, indata)

	log.Println("CustomerService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *customerBaseService) Delete(customerId string, delete_permanent bool) error {

	log.Println("CustomerService::Delete - Begin", customerId)

	if delete_permanent {
		result, err := p.daoCustomer.Delete(customerId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(customerId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CustomerService::Delete - End")
	return nil
}

// Authenticate - Authenticate User
func (p *customerBaseService) Authenticate(auth_key string, auth_login string, auth_pwd string) (utils.Map, error) {
	log.Println("Authenticate::  Begin ", auth_key, auth_login, auth_pwd)

	log.Println("User Password from API", auth_pwd)
	encpwd := utils.SHA(auth_pwd)
	dataUser, err := p.daoCustomer.Authenticate(auth_key, auth_login, encpwd)

	log.Println("Length of dataUser :", dataUser)

	if err != nil {
		err := &utils.AppError{ErrorCode: "S30340101", ErrorMsg: "Wrong Credentials", ErrorDetail: "Authenticate credentials is wrong !!"}
		return utils.Map{}, err
	}

	dataval, dataok := dataUser[db_common.FLD_IS_DELETED]
	if dataok && !dataval.(bool) {
		err := &utils.AppError{ErrorCode: "S30340102", ErrorMsg: "User not in Active Mode. Contact Admin!", ErrorDetail: "User not in Active Mode. Contact Admin!"}
		return utils.Map{}, err
	}

	dataval, dataok = dataUser[db_common.FLD_IS_VERIFIED]
	if dataok && !dataval.(bool) {
		err := &utils.AppError{ErrorCode: "S30340103", ErrorMsg: "User not yet verified!", ErrorDetail: "User not yet verified!!"}
		return utils.Map{}, err
	}

	return dataUser, nil
}

// Change Password - Change Customer Password
func (p *customerBaseService) ChangePassword(userid string, newpwd string) (utils.Map, error) {

	log.Println("AppUserService::ChangePassword - Begin")
	indata := utils.Map{
		sales_common.FLD_CUSTOMER_PASSWORD: utils.SHA(newpwd),
	}
	data, err := p.daoCustomer.Update(userid, indata)

	log.Println("AppUserService::ChangePassword - End ")
	return data, err
}

func (p *customerBaseService) errorReturn(err error) (CustomerService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
