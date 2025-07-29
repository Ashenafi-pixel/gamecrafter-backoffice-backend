CREATE TABLE banners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page VARCHAR(50) NOT NULL,
    page_url TEXT NOT NULL,
    image_url TEXT NOT NULL,
    headline VARCHAR(255) NOT NULL,
    tagline VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add unique constraint on page to prevent duplicate banners for the same page
ALTER TABLE banners ADD CONSTRAINT unique_page UNIQUE (page);