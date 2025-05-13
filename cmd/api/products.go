package main

import (
	"github.com/pistolricks/vault199-api/internal/data"
	"github.com/pistolricks/vault199-api/internal/validator"
	"net/http"
)

// Using the data.Product type instead of defining our own

func (app *application) productsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Style string
		Mill  string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Style = app.readString(qs, "attrs->>style", "")

	input.Mill = app.readString(qs, "attrs->>mill", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "attrs->>product_title", "category_name", "subcategory_name", "color_name", "sizes", "mill", "msrp", "suggested_price", "map_pricing", "-id", "-product_title", "-category_name", "-msrp", "-suggested_price", "-map_pricing", "-subcategory_name", "-color_name", "-size", "-mill"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	products, metadata, err := app.models.Products.GetAll(input.Style, input.Mill, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"products": products, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) importProducts(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Products []data.Attrs `json:"products"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	for _, res := range input.Products {

		attrs := new(data.Attrs)

		*attrs = res

		product := new(data.Product)

		product.ID = res.ID
		product.Type = "sanmar"
		product.Attrs = *attrs

		err = app.models.Products.Insert(product)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

	}

}
