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

// PreferenceService - Business Preference Service structure
type PreferenceService interface {
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
type preferenceBaseService struct {
	db_utils.DatabaseService
	dbRegion      db_utils.DatabaseService
	daoPreference sales_repository.PreferenceDao
	daoBusiness   platform_repository.BusinessDao
	child         PreferenceService
	businessId    string
}

// NewPreferenceService - Construct Preference
func NewPreferenceService(props utils.Map) (PreferenceService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("PreferenceService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := preferenceBaseService{}
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

// preferencesBaseService - Close all the services
func (p *preferenceBaseService) EndService() {
	log.Printf("EndPreferenceService ")
	p.CloseDatabaseService()
}

func (p *preferenceBaseService) initializeService() {
	log.Printf("PreferenceMongoService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoPreference = sales_repository.NewPreferenceDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *preferenceBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("PreferenceService::FindAll - Begin")

	listdata, err := p.daoPreference.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("PreferenceService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *preferenceBaseService) Get(preferenceId string) (utils.Map, error) {
	log.Printf("PreferenceService::Get::  Begin %v", preferenceId)

	data, err := p.daoPreference.Get(preferenceId)

	log.Println("PreferenceService::Get:: End ", err)
	return data, err
}

func (p *preferenceBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("PreferenceBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoPreference.Find(filter)
	log.Println("PreferenceBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *preferenceBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("PreferenceService::Create - Begin")
	var preferenceId string

	dataval, dataok := indata[sales_common.FLD_PREFERENCE_ID]
	if dataok {
		preferenceId = strings.ToLower(dataval.(string))
	} else {
		preferenceId = utils.GenerateUniqueId("pre")
		log.Println("Unique Preference ID", preferenceId)
	}

	//BusinessPreference
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_PREFERENCE_ID] = preferenceId

	data, err := p.daoPreference.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("PreferenceService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *preferenceBaseService) Update(preferenceId string, indata utils.Map) (utils.Map, error) {

	log.Println("BusinessPreferenceService::Update - Begin")

	data, err := p.daoPreference.Update(preferenceId, indata)

	log.Println("PreferenceService::Update - End")
	return data, err
}

// Delete - Delete Service
func (p *preferenceBaseService) Delete(preferenceId string, delete_permanent bool) error {

	log.Println("PreferenceService::Delete - Begin", preferenceId)

	if delete_permanent {
		result, err := p.daoPreference.Delete(preferenceId)
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

	log.Printf("PreferenceService::Delete - End")
	return nil
}

func (p *preferenceBaseService) errorReturn(err error) (PreferenceService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
