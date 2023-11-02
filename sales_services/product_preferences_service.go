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

// ProdPreferenceService - Business ProdPreference Service structure
type ProdPreferenceService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(preferenceId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(preferenceId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(preferenceId string, delete_permanent bool) error

	EndService()
}

// preferenceService - Business preference Service structure
type prodPreferenceBaseService struct {
	db_utils.DatabaseService
	dbRegion          db_utils.DatabaseService
	daoProdPreference sales_repository.ProdPreferenceDao
	daoBusiness       platform_repository.BusinessDao
	child             ProdPreferenceService
	businessId        string
}

// NewProdPreferenceService - Construct ProdPreference
func NewProdPreferenceService(props utils.Map) (ProdPreferenceService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("ProdPreferenceService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := prodPreferenceBaseService{}
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
			ErrorMsg:    "Invalid business_id",
			ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// preferencesBaseService - Close all the services
func (p *prodPreferenceBaseService) EndService() {
	log.Printf("EndProdPreferenceService ")
	p.CloseDatabaseService()
}

func (p *prodPreferenceBaseService) initializeService() {
	log.Printf("ProdPreferenceMongoService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoProdPreference = sales_repository.NewProdPreferenceDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *prodPreferenceBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("ProdPreferenceService::FindAll - Begin")

	listdata, err := p.daoProdPreference.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("ProdPreferenceService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *prodPreferenceBaseService) Get(preferenceId string) (utils.Map, error) {
	log.Printf("ProdPreferenceService::Get::  Begin %v", preferenceId)

	data, err := p.daoProdPreference.Get(preferenceId)

	log.Println("ProdPreferenceService::Get:: End ", err)
	return data, err
}

func (p *prodPreferenceBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("ProdPreferenceBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoProdPreference.Find(filter)
	log.Println("ProdPreferenceBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *prodPreferenceBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("ProdPreferenceService::Create - Begin")
	var preferenceId string

	dataval, dataok := indata[sales_common.FLD_PROD_PREFERENCE_ID]
	if dataok {
		preferenceId = strings.ToLower(dataval.(string))
	} else {
		preferenceId = utils.GenerateUniqueId("prodpre")
		log.Println("Unique ProdPreference ID", preferenceId)
	}

	//BusinessProdPreference
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_PROD_PREFERENCE_ID] = preferenceId

	data, err := p.daoProdPreference.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("ProdPreferenceService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *prodPreferenceBaseService) Update(preferenceId string, indata utils.Map) (utils.Map, error) {

	log.Println("BusinessProdPreferenceService::Update - Begin")

	data, err := p.daoProdPreference.Update(preferenceId, indata)

	log.Println("ProdPreferenceService::Update - End")
	return data, err
}

// Delete - Delete Service
func (p *prodPreferenceBaseService) Delete(preferenceId string, delete_permanent bool) error {

	log.Println("ProdPreferenceService::Delete - Begin", preferenceId)

	if delete_permanent {
		result, err := p.daoProdPreference.Delete(preferenceId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(preferenceId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("ProdPreferenceService::Delete - End")
	return nil
}

func (p *prodPreferenceBaseService) errorReturn(err error) (ProdPreferenceService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
