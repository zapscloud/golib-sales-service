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

// PoliciesService - Business Policies Service structure
type PoliciesService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(policyId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(policyId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(policyId string, delete_permanent bool) error

	EndService()
}

// policiesService - Business policies Service structure
type policiesBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoPolicies sales_repository.PoliciesDao
	daoBusiness platform_repository.BusinessDao
	child       PoliciesService
	businessId  string
}

// NewPoliciesService - Construct Policies
func NewPoliciesService(props utils.Map) (PoliciesService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("PoliciesService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := policiesBaseService{}
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

// policiesBaseService - Close all the services
func (p *policiesBaseService) EndService() {
	log.Printf("EndPoliciesService ")
	p.CloseDatabaseService()
}

func (p *policiesBaseService) initializeService() {
	log.Printf("PoliciesMongoService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoPolicies = sales_repository.NewPoliciesDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *policiesBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("PoliciesService::FindAll - Begin")

	listdata, err := p.daoPolicies.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("PoliciesService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *policiesBaseService) Get(policyId string) (utils.Map, error) {
	log.Printf("PoliciesService::Get::  Begin %v", policyId)

	data, err := p.daoPolicies.Get(policyId)

	log.Println("PoliciesService::Get:: End ", err)
	return data, err
}

func (p *policiesBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("PoliciesBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoPolicies.Find(filter)
	log.Println("PoliciesBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *policiesBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("PoliciesService::Create - Begin")
	var policyId string

	dataval, dataok := indata[sales_common.FLD_POLICY_ID]
	if dataok {
		policyId = strings.ToLower(dataval.(string))
	} else {
		policyId = utils.GenerateUniqueId("pol")
		log.Println("Unique Policies ID", policyId)
	}

	// Change the PolicyType to Uppercase
	dataval, dataok = indata[sales_common.FLD_POLICY_TYPE]
	if dataok {
		indata[sales_common.FLD_POLICY_TYPE] = strings.ToUpper(dataval.(string))
	}

	// Business Policies
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_POLICY_ID] = policyId

	data, err := p.daoPolicies.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("PoliciesService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *policiesBaseService) Update(policyId string, indata utils.Map) (utils.Map, error) {

	log.Println("BusinessPoliciesService::Update - Begin")

	data, err := p.daoPolicies.Update(policyId, indata)

	log.Println("PoliciesService::Update - End")
	return data, err
}

// Delete - Delete Service
func (p *policiesBaseService) Delete(policyId string, delete_permanent bool) error {

	log.Println("PoliciesService::Delete - Begin", policyId)

	if delete_permanent {
		result, err := p.daoPolicies.Delete(policyId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(policyId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("PoliciesService::Delete - End")
	return nil
}

func (p *policiesBaseService) errorReturn(err error) (PoliciesService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
