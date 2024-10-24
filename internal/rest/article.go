package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"

	"apismrtbiz/domain"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// ArticleService represent the article's usecases
//
//go:generate mockery --name ArticleService
type ArticleService interface {
	Fetch(ctx context.Context, cursor string, num int64) ([]domain.Article, string, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	Update(ctx context.Context, ar *domain.Article) error
	GetByTitle(ctx context.Context, title string) (domain.Article, error)
	Store(context.Context, *domain.Article) error
	Delete(ctx context.Context, id int64) error
}

// ArticleHandler  represent the httphandler for article
type ArticleHandler struct {
	Service ArticleService
}

const defaultNum = 10

// NewArticleHandler will initialize the articles/ resources endpoint
func NewArticleHandler(e *fiber.App, svc ArticleService) {
	handler := &ArticleHandler{
		Service: svc,
	}
	e.Get("/articles", handler.FetchArticle)
	e.Post("/articles", handler.Store)
	e.Get("/articles/:id", handler.GetByID)
	e.Delete("/articles/:id", handler.Delete)
}

// FetchArticle will fetch the article based on given params
func (a *ArticleHandler) FetchArticle(c *fiber.Ctx) error {

	numS := c.Query("num")
	num, err := strconv.Atoi(numS)
	if err != nil || num == 0 {
		num = defaultNum
	}

	cursor := c.Query("cursor")

	listAr, nextCursor, err := a.Service.Fetch(c.Context(), cursor, int64(num))

	if rep := ReturnErr(c, err); rep != nil {
		return rep
	}

	c.Set(`X-Cursor`, nextCursor)

	return c.JSON(listAr)
}

type errRep struct {
	Message string `json:"message,omitempty"`
}

func ReturnErr(c *fiber.Ctx, er error) error {
	var rep error
	if er != nil {
		rep = c.JSON(errRep{er.Error()})
	}
	return rep
}

// GetByID will get article by given id
func (a *ArticleHandler) GetByID(c *fiber.Ctx) error {
	idP, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
	}

	id := int64(idP)

	art, err := a.Service.GetByID(c.Context(), id)
	if rep := ReturnErr(c, err); rep != nil {
		return rep
	}

	return c.JSON(art)
}

func isRequestValid(m *domain.Article) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Store will store the article by given request body
func (a *ArticleHandler) Store(c *fiber.Ctx) (err error) {
	var article domain.Article

	//err = c.Bind(&article)
	//if err != nil {
	//	return c.JSON(http.StatusUnprocessableEntity, err.Error())
	//}
	//
	//var ok bool
	//if ok, err = isRequestValid(&article); !ok {
	//	return c.JSON(http.StatusBadRequest, err.Error())
	//}

	err = a.Service.Store(c.Context(), &article)
	if rep := ReturnErr(c, err); rep != nil {
		return rep
	}
	return c.JSON(article)
}

// Delete will delete article by given param
func (a *ArticleHandler) Delete(c *fiber.Ctx) error {
	idP, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
	}

	id := int64(idP)

	err = a.Service.Delete(c.Context(), id)
	if rep := ReturnErr(c, err); rep != nil {
		return rep
	}

	return nil
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
