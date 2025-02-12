package controller

import (
	"context"
	"product-crud/cache"
	"product-crud/constant/errorconstants"
	"product-crud/dto/app"
	"product-crud/dto/request"
	"product-crud/dto/response"
	"product-crud/service"
	"product-crud/util"
	"product-crud/util/apiresponse"
	"product-crud/util/logger"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type IProductController interface {
	GetAllProduct(c *gin.Context)
	GetProductById(c *gin.Context)
	AddProduct(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
}

type ProductController struct {
	productService service.IProductService
}

func NewProductController(productService service.IProductService) ProductController {
	logger.Info("Initializing product controller..")
	return ProductController{
		productService: productService,
	}
}

func (pc ProductController) GetAllProduct(c *gin.Context) {
	logger.Info("Get all product request")
	pagination, err := util.GeneratePaginationFromRequest(c)
	if err != nil {
		panic(err)
	}

	hash, err := util.HashFromStruct(pagination)
	if err != nil {
		panic(err)
	}
	key := "GetAllProduct:all:" + hash

	var products app.PaginatedResult[response.GetProductResponse]
	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()
	if c.DefaultQuery("no_cache", "0") == "0" {
		err := cache.Get(ctx, key, &products)
		if err != nil {
			logger.Error("Error : %v", err)
		}
	}

	isFromCache := false
	if !products.IsEmpty() {
		isFromCache = true
	} else {
		val, err := pc.productService.GetAll(pagination)
		if err != nil {
			panic(err)
		}
		products = val
		go func() {
			ctx, cancel := context.WithTimeout(c, 3*time.Second)
			defer cancel()
			err := cache.Set(ctx, key, products)
			if err != nil {
				logger.Error("Error : %v", err)
			}
		}()
	}

	logger.Info("Get all product success")
	apiresponse.Ok(c, products, isFromCache)
}

func (pc ProductController) GetProductById(c *gin.Context) {
	logger.Info(`Get product by id request, id = %s`, c.Param("id"))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Error : %v", err)
		panic(errorconstants.INTERNAL_ERROR)
	}

	key := "GetProductById:" + c.Param("id")

	var product response.GetProductResponse
	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()
	if c.DefaultQuery("no_cache", "0") == "0" {
		err := cache.Get(ctx, key, &product)
		if err != nil {
			logger.Error("Error : %v", err)
		}
	}

	isFromCache := false
	if !product.IsEmpty() {
		isFromCache = true
	} else {
		val, err := pc.productService.GetById(uint(id))
		if err != nil {
			panic(err)
		}
		product = val
		go func() {
			ctx, cancel := context.WithTimeout(c, 3*time.Second)
			defer cancel()
			err := cache.Set(ctx, key, product)
			if err != nil {
				logger.Error("Error : %v", err)
			}
		}()
	}

	logger.Info(`Get product by id, id = %s success`, c.Param("id"))
	apiresponse.Ok(c, product, isFromCache)
}

func (pc ProductController) AddProduct(c *gin.Context) {
	logger.Info(`Add new product request`)
	var request request.ProductAddRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		panic(err)
	}
	user, err := util.GetUserClaims(c)
	if err != nil {
		panic(err)
	}
	product, err := pc.productService.AddProduct(request, user.UserId)
	if err != nil {
		panic(err)
	}

	logger.Info(`Add new product success`)
	apiresponse.Ok(c, product, false)
}

func (pc ProductController) UpdateProduct(c *gin.Context) {
	logger.Info(`Update product of id = %s`, c.Param("id"))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Error : %v", err)
		panic(errorconstants.INTERNAL_ERROR)
	}
	var request request.ProductUpdateRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		panic(err)
	}
	user, err := util.GetUserClaims(c)
	if err != nil {
		panic(err)
	}
	product, err := pc.productService.UpdateProduct(uint(id), request, user.UserId)
	if err != nil {
		panic(err)
	}
	logger.Info(`Update product of id = %s success`, c.Param("id"))
	apiresponse.Ok(c, product, false)
}

func (pc ProductController) DeleteProduct(c *gin.Context) {
	logger.Info(`Delete product of id = %s`, c.Param("id"))
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Error : %v", err)
		panic(errorconstants.INTERNAL_ERROR)
	}
	user, err := util.GetUserClaims(c)
	if err != nil {
		panic(err)
	}
	err = pc.productService.DeleteProduct(uint(id), user.UserId)
	if err != nil {
		panic(err)
	}
	logger.Info(`Delete product of id = %s success`, c.Param("id"))
	apiresponse.Ok(c, nil, false)
}

var _ IProductController = (*ProductController)(nil)
