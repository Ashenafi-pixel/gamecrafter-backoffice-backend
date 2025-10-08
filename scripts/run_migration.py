#!/usr/bin/env python3
"""
SQL Migration Runner for AWS Server
Date: 2025-10-08
Description: Run SQL migration script on PostgreSQL database
"""

import psycopg2
import sys
import os

# Database connection parameters - UPDATE THESE FOR YOUR AWS SERVER
DB_CONFIG = {
    'host': 'your-aws-rds-endpoint.amazonaws.com',  # Update with your RDS endpoint
    'port': '5432',  # Standard PostgreSQL port
    'database': 'tucanbit',
    'user': 'your-db-username',  # Update with your DB username
    'password': 'your-db-password'  # Update with your DB password
}

def run_migration():
    """Run the SQL migration script"""
    
    migration_file = 'migrations/2025-10-08_schema_updates.sql'
    
    if not os.path.exists(migration_file):
        print(f"❌ Migration file not found: {migration_file}")
        sys.exit(1)
    
    try:
        # Connect to database
        print(f"🔌 Connecting to database: {DB_CONFIG['host']}:{DB_CONFIG['port']}")
        conn = psycopg2.connect(**DB_CONFIG)
        conn.autocommit = True  # Enable autocommit for DDL statements
        cursor = conn.cursor()
        
        print(f"📁 Reading migration file: {migration_file}")
        
        # Read and execute the migration file
        with open(migration_file, 'r', encoding='utf-8') as file:
            migration_sql = file.read()
        
        print("🚀 Executing migration...")
        cursor.execute(migration_sql)
        
        print("✅ Migration completed successfully!")
        
        # Run verification queries
        print("\n🔍 Running verification queries...")
        
        # Check games table structure
        cursor.execute("""
            SELECT column_name, data_type, is_nullable 
            FROM information_schema.columns 
            WHERE table_name = 'games' 
            ORDER BY ordinal_position
        """)
        games_columns = cursor.fetchall()
        print(f"📊 Games table columns: {len(games_columns)}")
        
        # Check groove_transactions table structure
        cursor.execute("""
            SELECT column_name, data_type, is_nullable 
            FROM information_schema.columns 
            WHERE table_name = 'groove_transactions' 
            AND column_name IN ('balance_before', 'balance_after')
        """)
        balance_columns = cursor.fetchall()
        print(f"💰 Balance columns added: {len(balance_columns)}")
        
        # Check if Sweet Bonanza was added
        cursor.execute("SELECT * FROM games WHERE game_id = '82695'")
        sweet_bonanza = cursor.fetchone()
        if sweet_bonanza:
            print("🎮 Sweet Bonanza (82695): ✅ Found")
        else:
            print("🎮 Sweet Bonanza (82695): ❌ Not found")
        
        # Count total games
        cursor.execute("SELECT COUNT(*) FROM games")
        total_games = cursor.fetchone()[0]
        print(f"📈 Total games in database: {total_games}")
        
    except psycopg2.Error as e:
        print(f"❌ Database error: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"❌ Migration failed: {e}")
        sys.exit(1)
    finally:
        if 'cursor' in locals():
            cursor.close()
        if 'conn' in locals():
            conn.close()

if __name__ == "__main__":
    print("🗃️  SQL Migration Runner")
    print("=" * 50)
    run_migration()
    print("\n🎉 Migration process completed!")
