#!/usr/bin/env python3
"""
Database Test Script for TucanBIT Online Casino
This script connects to the PostgreSQL database and lists all users.
"""

import psycopg2
import json
from datetime import datetime

# Database configuration
DB_CONFIG = {
    'host': 'localhost',
    'port': 5433,
    'database': 'tucanbit',
    'user': 'tucanbit',
    'password': '5kj0YmV5FKKpU9D50B7yH5A'
}

def test_connection():
    """Test database connection"""
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        print("✅ Database connection successful!")
        return conn
    except Exception as e:
        print(f"❌ Database connection failed: {e}")
        return None

def list_tables(conn):
    """List all tables in the database"""
    try:
        cursor = conn.cursor()
        cursor.execute("""
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public'
            ORDER BY table_name;
        """)
        tables = cursor.fetchall()
        print("\n📋 Available Tables:")
        for table in tables:
            print(f"  - {table[0]}")
        cursor.close()
        return [table[0] for table in tables]
    except Exception as e:
        print(f"❌ Error listing tables: {e}")
        return []

def list_users(conn):
    """List all users from the users table"""
    try:
        cursor = conn.cursor()
        cursor.execute("""
            SELECT id, phone_number, first_name, last_name, email, created_at, status
            FROM users
            ORDER BY created_at DESC;
        """)
        users = cursor.fetchall()
        
        print(f"\n👥 Found {len(users)} users:")
        print("-" * 80)
        
        for user in users:
            user_id, phone, first_name, last_name, email, created_at, status = user
            print(f"ID: {user_id}")
            print(f"Phone: {phone}")
            print(f"Name: {first_name or 'N/A'} {last_name or 'N/A'}")
            print(f"Email: {email or 'N/A'}")
            print(f"Status: {status or 'N/A'}")
            print(f"Created: {created_at}")
            print("-" * 40)
        
        cursor.close()
        return users
    except Exception as e:
        print(f"❌ Error listing users: {e}")
        return []

def get_user_count(conn):
    """Get total user count"""
    try:
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM users;")
        count = cursor.fetchone()[0]
        cursor.close()
        return count
    except Exception as e:
        print(f"❌ Error getting user count: {e}")
        return 0

def get_recent_registrations(conn, limit=5):
    """Get recent user registrations"""
    try:
        cursor = conn.cursor()
        cursor.execute("""
            SELECT id, phone_number, first_name, last_name, created_at
            FROM users
            ORDER BY created_at DESC
            LIMIT %s;
        """, (limit,))
        users = cursor.fetchall()
        
        print(f"\n🆕 Recent {len(users)} Registrations:")
        print("-" * 60)
        
        for user in users:
            user_id, phone, first_name, last_name, created_at = user
            print(f"📱 {phone} - {first_name or 'N/A'} {last_name or 'N/A'}")
            print(f"   ID: {user_id} | Created: {created_at}")
            print()
        
        cursor.close()
        return users
    except Exception as e:
        print(f"❌ Error getting recent registrations: {e}")
        return []

def main():
    """Main function"""
    print("🎰 TucanBIT Online Casino - Database Test")
    print("=" * 50)
    
    # Test connection
    conn = test_connection()
    if not conn:
        return
    
    try:
        # List tables
        tables = list_tables(conn)
        
        # Check if users table exists
        if 'users' in tables:
            # Get user count
            count = get_user_count(conn)
            print(f"\n📊 Total Users: {count}")
            
            # List all users
            users = list_users(conn)
            
            # Get recent registrations
            recent = get_recent_registrations(conn, 10)
            
        else:
            print("\n⚠️  Users table not found. Available tables:")
            for table in tables:
                print(f"  - {table}")
    
    finally:
        conn.close()
        print("\n🔌 Database connection closed.")

if __name__ == "__main__":
    main() 