CREATE INDEX IF NOT EXISTS idx_products_attrs ON products USING gin (attrs);
CREATE INDEX IF NOT EXISTS idx_products_attrs_style ON products USING gin ((attrs->'style'));

