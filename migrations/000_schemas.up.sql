-- Базовые схемы для разделения данных по сервисам
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS customers;
CREATE SCHEMA IF NOT EXISTS vehicles;
CREATE SCHEMA IF NOT EXISTS deals;
CREATE SCHEMA IF NOT EXISTS parts;
CREATE SCHEMA IF NOT EXISTS brands;
CREATE SCHEMA IF NOT EXISTS dealerpoints;

-- Если таблицы уже были в public, переносим их в целевые схемы.
DO $$
BEGIN
    IF to_regclass('auth.users') IS NULL AND to_regclass('public.users') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.users SET SCHEMA auth';
    END IF;
    IF to_regclass('auth.roles') IS NULL AND to_regclass('public.roles') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.roles SET SCHEMA auth';
    END IF;
    IF to_regclass('customers.customers') IS NULL AND to_regclass('public.customers') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.customers SET SCHEMA customers';
    END IF;
    IF to_regclass('vehicles.vehicles') IS NULL AND to_regclass('public.vehicles') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.vehicles SET SCHEMA vehicles';
    END IF;
    IF to_regclass('deals.deals') IS NULL AND to_regclass('public.deals') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.deals SET SCHEMA deals';
    END IF;
    IF to_regclass('parts.parts') IS NULL AND to_regclass('public.parts') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.parts SET SCHEMA parts';
    END IF;
    IF to_regclass('parts.part_folders') IS NULL AND to_regclass('public.part_folders') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.part_folders SET SCHEMA parts';
    END IF;
    IF to_regclass('parts.part_stock') IS NULL AND to_regclass('public.part_stock') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.part_stock SET SCHEMA parts';
    END IF;
    IF to_regclass('brands.brands') IS NULL AND to_regclass('public.brands') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.brands SET SCHEMA brands';
    END IF;
    IF to_regclass('dealerpoints.dealer_points') IS NULL AND to_regclass('public.dealer_points') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.dealer_points SET SCHEMA dealerpoints';
    END IF;
    IF to_regclass('dealerpoints.legal_entities') IS NULL AND to_regclass('public.legal_entities') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.legal_entities SET SCHEMA dealerpoints';
    END IF;
    IF to_regclass('dealerpoints.dealer_point_legal_entities') IS NULL AND to_regclass('public.dealer_point_legal_entities') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.dealer_point_legal_entities SET SCHEMA dealerpoints';
    END IF;
    IF to_regclass('dealerpoints.warehouses') IS NULL AND to_regclass('public.warehouses') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE public.warehouses SET SCHEMA dealerpoints';
    END IF;
END $$;
