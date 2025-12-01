#!/usr/bin/env python3
"""
Script to generate seed data INSERT statements for:
- Permissions (204 records)
- Role Permissions (203 records for admin role)
- Pages (31 records)
- Admin Activity Actions (24 records)

Usage:
    python3 scripts/generate_seed_data.py > migrations/seed_additional_data.sql
"""

import os
import sys
import subprocess
import json

# Database connection from config
DB_HOST = "localhost"
DB_PORT = "5433"
DB_NAME = "tucanbit"
DB_USER = "tucanbit"
DB_PASSWORD = "5kj0YmV5FKKpU9D50B7yH5A"

def run_query(query):
    """Execute a PostgreSQL query and return results"""
    env = os.environ.copy()
    env['PGPASSWORD'] = DB_PASSWORD
    
    cmd = [
        'psql',
        '-h', DB_HOST,
        '-p', DB_PORT,
        '-U', DB_USER,
        '-d', DB_NAME,
        '-t', '-A', '-F', '|'
    ]
    
    try:
        result = subprocess.run(
            cmd,
            input=query,
            text=True,
            capture_output=True,
            env=env,
            check=True
        )
        return result.stdout.strip().split('\n') if result.stdout.strip() else []
    except subprocess.CalledProcessError as e:
        print(f"Error executing query: {e}", file=sys.stderr)
        print(f"Error output: {e.stderr}", file=sys.stderr)
        return []

def escape_sql_string(value):
    """Escape SQL string values"""
    if value is None:
        return 'NULL'
    if isinstance(value, str):
        return "'" + value.replace("'", "''") + "'"
    return str(value)

def generate_permissions_seed():
    """Generate INSERT statements for permissions"""
    print("-- ============================================================================")
    print("-- 11. PERMISSIONS")
    print("-- ============================================================================")
    
    # Check if permissions table has resource/action columns
    check_query = """
    SELECT column_name 
    FROM information_schema.columns 
    WHERE table_name = 'permissions' AND column_name IN ('resource', 'action')
    """
    columns = run_query(check_query)
    has_resource_action = len(columns) >= 2
    
    if has_resource_action:
        query = "SELECT id, name, description, resource, action FROM permissions ORDER BY resource, action;"
    else:
        query = "SELECT id, name, description FROM permissions ORDER BY name;"
    
    rows = run_query(query)
    
    if not rows or rows == ['']:
        print("-- No permissions found")
        return
    
    print("INSERT INTO permissions (id, name, description" + (", resource, action" if has_resource_action else "") + ") VALUES")
    
    values = []
    for row in rows:
        if not row.strip():
            continue
        parts = row.split('|')
        if has_resource_action and len(parts) >= 5:
            perm_id, name, desc, resource, action = parts[0], parts[1], parts[2] if len(parts) > 2 else '', parts[3] if len(parts) > 3 else '', parts[4] if len(parts) > 4 else ''
            values.append(f"({escape_sql_string(perm_id)}, {escape_sql_string(name)}, {escape_sql_string(desc)}, {escape_sql_string(resource)}, {escape_sql_string(action)})")
        elif len(parts) >= 3:
            perm_id, name, desc = parts[0], parts[1], parts[2] if len(parts) > 2 else ''
            values.append(f"({escape_sql_string(perm_id)}, {escape_sql_string(name)}, {escape_sql_string(desc)})")
    
    print(",\n".join(values) + "\nON CONFLICT (id) DO NOTHING;")
    print()

def generate_role_permissions_seed():
    """Generate INSERT statements for role_permissions"""
    print("-- ============================================================================")
    print("-- 12. ROLE PERMISSIONS (Admin Role)")
    print("-- ============================================================================")
    
    admin_role_id = "33dbb86c-e306-4d1d-b7df-cdf556e1ae32"
    
    query = f"""
    SELECT rp.id, rp.role_id, rp.permission_id, rp.value 
    FROM role_permissions rp 
    WHERE rp.role_id = '{admin_role_id}' 
    ORDER BY rp.permission_id;
    """
    
    rows = run_query(query)
    
    if not rows or rows == ['']:
        print("-- No role permissions found")
        return
    
    print("INSERT INTO role_permissions (id, role_id, permission_id, value) VALUES")
    
    values = []
    for row in rows:
        if not row.strip():
            continue
        parts = row.split('|')
        if len(parts) >= 4:
            rp_id, role_id, perm_id, value = parts[0], parts[1], parts[2], parts[3] if len(parts) > 3 and parts[3] else 'NULL'
            value_str = value if value and value != '' else 'NULL'
            values.append(f"({escape_sql_string(rp_id)}, {escape_sql_string(role_id)}, {escape_sql_string(perm_id)}, {value_str})")
    
    print(",\n".join(values) + "\nON CONFLICT (id) DO NOTHING;")
    print()

def generate_pages_seed():
    """Generate INSERT statements for pages"""
    print("-- ============================================================================")
    print("-- 13. PAGES")
    print("-- ============================================================================")
    
    query = "SELECT id, path, label, parent_id, icon, created_at, updated_at FROM pages ORDER BY path;"
    
    rows = run_query(query)
    
    if not rows or rows == ['']:
        print("-- No pages found")
        return
    
    print("INSERT INTO pages (id, path, label, parent_id, icon, created_at, updated_at) VALUES")
    
    values = []
    for row in rows:
        if not row.strip():
            continue
        parts = row.split('|')
        if len(parts) >= 7:
            page_id, path, label, parent_id, icon, created_at, updated_at = parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6]
            parent_id_str = escape_sql_string(parent_id) if parent_id and parent_id.strip() else 'NULL'
            icon_str = escape_sql_string(icon) if icon and icon.strip() else 'NULL'
            values.append(f"({escape_sql_string(page_id)}, {escape_sql_string(path)}, {escape_sql_string(label)}, {parent_id_str}, {icon_str}, {escape_sql_string(created_at)}, {escape_sql_string(updated_at)})")
    
    print(",\n".join(values) + "\nON CONFLICT (id) DO NOTHING;")
    print()

def generate_admin_activity_actions_seed():
    """Generate INSERT statements for admin_activity_actions"""
    print("-- ============================================================================")
    print("-- 14. ADMIN ACTIVITY ACTIONS")
    print("-- ============================================================================")
    
    query = "SELECT id, name, description, category_id, is_active, created_at FROM admin_activity_actions ORDER BY name;"
    
    rows = run_query(query)
    
    if not rows or rows == ['']:
        print("-- No admin activity actions found")
        return
    
    print("INSERT INTO admin_activity_actions (id, name, description, category_id, is_active, created_at) VALUES")
    
    values = []
    for row in rows:
        if not row.strip():
            continue
        parts = row.split('|')
        if len(parts) >= 6:
            action_id, name, desc, category_id, is_active, created_at = parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]
            desc_str = escape_sql_string(desc) if desc and desc.strip() else 'NULL'
            category_id_str = escape_sql_string(category_id) if category_id and category_id.strip() else 'NULL'
            is_active_bool = 'true' if is_active and is_active.lower() in ('t', 'true', '1') else 'false'
            values.append(f"({escape_sql_string(action_id)}, {escape_sql_string(name)}, {desc_str}, {category_id_str}, {is_active_bool}, {escape_sql_string(created_at)})")
    
    print(",\n".join(values) + "\nON CONFLICT (id) DO NOTHING;")
    print()

def main():
    print("-- Additional Seed Data")
    print("-- Generated by scripts/generate_seed_data.py")
    print("-- This file should be appended to the seed migration or run separately")
    print()
    
    generate_permissions_seed()
    generate_role_permissions_seed()
    generate_pages_seed()
    generate_admin_activity_actions_seed()
    
    print("-- End of additional seed data")

if __name__ == "__main__":
    main()

