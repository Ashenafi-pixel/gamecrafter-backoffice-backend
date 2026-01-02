-- Create role_allowed_pages table for role-based page access
CREATE TABLE IF NOT EXISTS public.role_allowed_pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role_id uuid NOT NULL,
    page_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT role_allowed_pages_pkey PRIMARY KEY (id),
    CONSTRAINT uq_role_allowed_pages UNIQUE (role_id, page_id),
    CONSTRAINT fk_role_allowed_pages_role FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_allowed_pages_page FOREIGN KEY (page_id) REFERENCES public.pages(id) ON DELETE CASCADE
);

COMMENT ON TABLE public.role_allowed_pages IS 'Junction table mapping roles to their allowed pages';

CREATE INDEX IF NOT EXISTS idx_role_allowed_pages_role_id ON public.role_allowed_pages(role_id);
CREATE INDEX IF NOT EXISTS idx_role_allowed_pages_page_id ON public.role_allowed_pages(page_id);

