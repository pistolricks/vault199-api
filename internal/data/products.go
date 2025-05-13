package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type NullString struct {
	sql.NullString
}

func (s NullString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.String)
}

func (s *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		s.String, s.Valid = "", false
		return nil
	}
	s.String, s.Valid = string(data), true
	return nil
}

type Product struct {
	ID    int64
	Type  string
	Attrs Attrs
}

type Attrs struct {
	ID                         int64      `json:"id"`
	ProductTitle               string     `json:"product_title"`
	ProductDescription         string     `json:"product_description"`
	Style                      string     `json:"style"`
	AvailableSizes             string     `json:"available_sizes"`
	BrandLogoImage             string     `json:"brand_logo_image"`
	ThumbnailImage             string     `json:"thumbnail_image"`
	ColorSwatchImage           string     `json:"color_swatch_image"`
	ProductImage               string     `json:"product_image"`
	SpecSheet                  string     `json:"spec_sheet"`
	PriceText                  string     `json:"price_text"`
	SuggestedPrice             string     `json:"suggested_price"`
	CategoryName               string     `json:"category_name"`
	SubcategoryName            string     `json:"subcategory_name"`
	ColorName                  string     `json:"color_name"`
	ColorSquareImage           string     `json:"color_square_image"`
	ColorProductImage          string     `json:"color_product_image"`
	ColorProductImageThumbnail string     `json:"color_product_image_thumbnail"`
	Size                       string     `json:"size"`
	PieceWeight                string     `json:"piece_weight"`
	PiecePrice                 string     `json:"piece_price"`
	DozensPrice                string     `json:"dozens_price"`
	CasePrice                  string     `json:"case_price"`
	PriceGroup                 string     `json:"price_group"`
	CaseSize                   string     `json:"case_size"`
	InventoryKey               string     `json:"inventory_key"`
	SizeIndex                  string     `json:"size_index"`
	SanmarMainframeColor       string     `json:"sanmar_mainframe_color"`
	Mill                       string     `json:"mill"`
	ProductStatus              string     `json:"product_status"`
	CompanionStyle             NullString `json:"companion_style"`
	Msrp                       string     `json:"msrp"`
	MapPricing                 NullString `json:"map_pricing,omitempty"`
	FrontModelImageUrl         string     `json:"front_model_image_url"`
	BackModelImageUrl          string     `json:"back_model_image_url"`
	FrontFlatImageUrl          string     `json:"front_flat_image_url"`
	BackFlatImageUrl           string     `json:"back_flat_image_url"`
	ProductMeasurements        string     `json:"product_measurements"`
	PmsColor                   NullString `json:"pms_color"`
	Gtin                       string     `json:"gtin"`
	DecorationSpecSheet        string     `json:"decoration_spec_sheet"`
}

type ProductModel struct {
	DB *sql.DB
}

func (m ProductModel) Insert(p *Product) error {

	attrsJSON, err := json.Marshal(p.Attrs)
	if err != nil {
		return err
	}

	err = m.DB.QueryRow(
		"INSERT INTO products (id, type, attrs) VALUES($1, $2, $3) RETURNING id",
		p.ID, p.Type, attrsJSON,
	).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil

	// Update the product with the new data from the database

}

func (m ProductModel) GetAll(style string, mill string, filters Filters) ([]*Product, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), *
FROM products
WHERE (to_tsvector('simple', attrs->>'style') @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (to_tsvector('simple', attrs->>'mill') @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY %s %s, id ASC
 LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{style, mill, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	products := []*Product{}

	for rows.Next() {
		var product Product
		var attrsJSON []byte

		err := rows.Scan(&totalRecords, &product.ID, &product.Type, &attrsJSON)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Unmarshal the JSON data into the Attrs struct
		err = json.Unmarshal(attrsJSON, &product.Attrs)
		if err != nil {
			return nil, Metadata{}, err
		}

		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return products, metadata, nil
}

func (m ProductModel) GetDistinctAll(style string, mill string, filters Filters) ([]*Product, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), p.*
FROM (
    SELECT DISTINCT ON (attrs->>'style') *
    FROM products
    WHERE (to_tsvector('simple', attrs->>'style') @@ plainto_tsquery('simple', $1) OR $1 = '')
    AND (to_tsvector('simple', attrs->>'mill') @@ plainto_tsquery('simple', $2) OR $2 = '')
) p
ORDER BY %s %s, id ASC
 LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{style, mill, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	products := []*Product{}

	for rows.Next() {
		var product Product
		var attrsJSON []byte

		err := rows.Scan(&totalRecords, &product.ID, &product.Type, &attrsJSON)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Unmarshal the JSON data into the Attrs struct
		err = json.Unmarshal(attrsJSON, &product.Attrs)
		if err != nil {
			return nil, Metadata{}, err
		}

		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return products, metadata, nil
}
