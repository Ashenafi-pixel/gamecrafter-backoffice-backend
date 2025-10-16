#!/usr/bin/env python3
"""
Game Data Import Script for AWS Server
Date: 2025-10-08
Description: Import game data from CSV file to PostgreSQL database
"""

import psycopg2
import csv
import os
import sys
from typing import Dict, Any

# Database connection parameters - UPDATE THESE FOR YOUR AWS SERVER
DB_CONFIG = {
    'host': 'localhost',  # Update with your RDS endpoint
    'port': '5433',  # Standard PostgreSQL port
    'database': 'tucanbit',
    'user': 'tucanbit',  # Update with your DB username
    'password': '5kj0YmV5FKKpU9D50B7yH5A'  # Update with your DB password
}

def get_db_connection() -> psycopg2.extensions.connection:
    """Get database connection with error handling"""
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        print(f"✅ Connected to database: {DB_CONFIG['host']}:{DB_CONFIG['port']}")
        return conn
    except psycopg2.Error as e:
        print(f"❌ Database connection failed: {e}")
        sys.exit(1)

def import_games_from_csv(csv_file_path: str) -> None:
    """Import games from CSV file to database"""
    
    if not os.path.exists(csv_file_path):
        print(f"❌ CSV file not found: {csv_file_path}")
        print("Please ensure the 'game 20250930-905.csv' file is in the same directory")
        sys.exit(1)
    
    conn = get_db_connection()
    cursor = conn.cursor()
    
    try:
        imported_count = 0
        updated_count = 0
        error_count = 0
        
        print(f"📁 Reading CSV file: {csv_file_path}")
        
        with open(csv_file_path, 'r', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            total_rows = sum(1 for _ in reader)
            file.seek(0)  # Reset file pointer
            reader = csv.DictReader(file)
            
            print(f"📊 Total rows to process: {total_rows}")
            
            for row_num, row in enumerate(reader, 1):
                try:
                    # Insert or update game record
                    cursor.execute("""
                        INSERT INTO games (game_id, internal_name, provider, integration_partner, name)
                        VALUES (%s, %s, %s, %s, %s)
                        ON CONFLICT (game_id) DO UPDATE SET
                            internal_name = EXCLUDED.internal_name,
                            provider = EXCLUDED.provider,
                            integration_partner = EXCLUDED.integration_partner,
                            name = EXCLUDED.name
                    """, (
                        row.get('game_id', ''),
                        row.get('internal_name', ''),
                        row.get('provider', ''),
                        'groovetech',  # Set integration_partner to groovetech
                        row.get('name', '')
                    ))
                    
                    # Check if it was an insert or update
                    if cursor.rowcount == 1:
                        imported_count += 1
                    else:
                        updated_count += 1
                        
                    # Progress indicator
                    if row_num % 100 == 0:
                        print(f"📈 Processed {row_num}/{total_rows} rows...")
                        
                except Exception as e:
                    error_count += 1
                    print(f"⚠️  Error processing row {row_num}: {e}")
                    continue
        
        # Commit all changes
        conn.commit()
        
        print(f"\n✅ Import completed successfully!")
        print(f"📊 Statistics:")
        print(f"   - New games imported: {imported_count}")
        print(f"   - Existing games updated: {updated_count}")
        print(f"   - Errors encountered: {error_count}")
        print(f"   - Total processed: {imported_count + updated_count}")
        
    except Exception as e:
        print(f"❌ Import failed: {e}")
        conn.rollback()
        sys.exit(1)
    finally:
        cursor.close()
        conn.close()

def verify_import() -> None:
    """Verify the import was successful"""
    conn = get_db_connection()
    cursor = conn.cursor()
    
    try:
        # Count total games
        cursor.execute("SELECT COUNT(*) FROM games")
        total_games = cursor.fetchone()[0]
        
        # Count groovetech games
        cursor.execute("SELECT COUNT(*) FROM games WHERE integration_partner = 'groovetech'")
        groovetech_games = cursor.fetchone()[0]
        
        # Check specific game
        cursor.execute("SELECT * FROM games WHERE game_id = '82695'")
        sweet_bonanza = cursor.fetchone()
        
        print(f"\n🔍 Verification Results:")
        print(f"   - Total games in database: {total_games}")
        print(f"   - GrooveTech games: {groovetech_games}")
        
        if sweet_bonanza:
            print(f"   - Sweet Bonanza (82695): ✅ Found")
        else:
            print(f"   - Sweet Bonanza (82695): ❌ Not found")
            
    except Exception as e:
        print(f"❌ Verification failed: {e}")
    finally:
        cursor.close()
        conn.close()

def main():
    """Main function"""
    print("🎮 Game Data Import Script")
    print("=" * 50)
    
    # Check if CSV file exists
    csv_file = "game 20250930-905.csv"
    
    if len(sys.argv) > 1:
        csv_file = sys.argv[1]
    
    print(f"📋 Configuration:")
    print(f"   - CSV file: {csv_file}")
    print(f"   - Database: {DB_CONFIG['host']}:{DB_CONFIG['port']}")
    print(f"   - Integration partner: groovetech")
    print()
    
    # Import games
    import_games_from_csv(csv_file)
    
    # Verify import
    verify_import()
    
    print("\n🎉 Import process completed!")

if __name__ == "__main__":
    main()
