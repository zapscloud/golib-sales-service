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

type RatingsService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(ratingId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(ratingId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(ratingId string, delete_permanent bool) error

	EndService()
}

type ratingsBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoRatings  sales_repository.RatingsDao
	daoBusiness platform_repository.BusinessDao
	child       RatingsService
	businessId  string
}

// NewRatingsService - Construct Ratings
func NewRatingsService(props utils.Map) (RatingsService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("RatingsService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := ratingsBaseService{}
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

// ratingsBaseService - Close all the services
func (p *ratingsBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *ratingsBaseService) initializeService() {
	log.Printf("RatingsService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoRatings = sales_repository.NewRatingsDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *ratingsBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("ratingsBaseService::FindAll - Begin")

	listdata, err := p.daoRatings.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("ratingsBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *ratingsBaseService) Get(ratingId string) (utils.Map, error) {
	log.Printf("ratingsBaseService::Get::  Begin %v", ratingId)

	data, err := p.daoRatings.Get(ratingId)

	log.Println("ratingsBaseService::Get:: End ", err)
	return data, err
}

func (p *ratingsBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("RatingsService::FindByCode::  Begin ", filter)

	data, err := p.daoRatings.Find(filter)
	log.Println("RatingsService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *ratingsBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("RatingsService::Create - Begin")
	var ratingId string

	dataval, dataok := indata[sales_common.FLD_RATING_ID]
	if dataok {
		ratingId = strings.ToLower(dataval.(string))
	} else {
		ratingId = utils.GenerateUniqueId("rati")
		log.Println("Unique Ratings ID", ratingId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_RATING_ID] = ratingId

	data, err := p.daoRatings.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("RatingsService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *ratingsBaseService) Update(ratingId string, indata utils.Map) (utils.Map, error) {

	log.Println("RatingsService::Update - Begin")

	data, err := p.daoRatings.Update(ratingId, indata)

	log.Println("RatingsService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *ratingsBaseService) Delete(ratingId string, delete_permanent bool) error {

	log.Println("RatingsService::Delete - Begin", ratingId)

	if delete_permanent {
		result, err := p.daoRatings.Delete(ratingId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(ratingId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("RatingsService::Delete - End")
	return nil
}

func (p *ratingsBaseService) errorReturn(err error) (RatingsService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
