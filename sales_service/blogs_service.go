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

type BlogService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(blogId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(blogId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(blogId string, delete_permanent bool) error

	EndService()
}

type blogBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoBlog     sales_repository.BlogDao
	daoBusiness platform_repository.BusinessDao
	child       BlogService
	businessId  string
}

// NewBlogService - Construct Blog
func NewBlogService(props utils.Map) (BlogService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("BlogService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := blogBaseService{}
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

// blogBaseService - Close all the services
func (p *blogBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *blogBaseService) initializeService() {
	log.Printf("BlogService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoBlog = sales_repository.NewBlogDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *blogBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("blogBaseService::FindAll - Begin")

	listdata, err := p.daoBlog.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("blogBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *blogBaseService) Get(blogId string) (utils.Map, error) {
	log.Printf("blogBaseService::Get::  Begin %v", blogId)

	data, err := p.daoBlog.Get(blogId)

	log.Println("blogBaseService::Get:: End ", err)
	return data, err
}

func (p *blogBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("BlogService::FindByCode::  Begin ", filter)

	data, err := p.daoBlog.Find(filter)
	log.Println("BlogService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *blogBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("BlogService::Create - Begin")
	var blogId string

	dataval, dataok := indata[sales_common.FLD_BLOG_ID]
	if dataok {
		blogId = strings.ToLower(dataval.(string))
	} else {
		blogId = utils.GenerateUniqueId("blo")
		log.Println("Unique Blog ID", blogId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_BLOG_ID] = blogId

	data, err := p.daoBlog.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("BlogService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *blogBaseService) Update(blogId string, indata utils.Map) (utils.Map, error) {

	log.Println("BlogService::Update - Begin")

	data, err := p.daoBlog.Update(blogId, indata)

	log.Println("BlogService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *blogBaseService) Delete(blogId string, delete_permanent bool) error {

	log.Println("BlogService::Delete - Begin", blogId)

	if delete_permanent {
		result, err := p.daoBlog.Delete(blogId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(blogId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("BlogService::Delete - End")
	return nil
}

func (p *blogBaseService) errorReturn(err error) (BlogService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
