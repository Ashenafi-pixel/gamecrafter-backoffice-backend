#!/usr/bin/env python3
"""
Corrected Game Import Script for AWS Server
Handles CSV with correct column names
"""

import csv
import psycopg2
import sys
from datetime import datetime

# Database connection parameters for AWS Docker
DB_CONFIG = {
    'host': 'localhost',
    'port': '5433',
    'database': 'tucanbit',
    'user': 'tucanbit',
    'password': '5kj0YmV5FKKpU9D50B7yH5A'
}

def import_games():
    """Import all games from CSV file"""
    print("🚀 Starting game import process...")
    
    try:
        # Connect to database
        conn = psycopg2.connect(**DB_CONFIG)
        cursor = conn.cursor()
        print("✅ Connected to database successfully")
        
        # Read CSV file
        csv_file = 'game 20250930-905.csv'
        imported_count = 0
        updated_count = 0
        error_count = 0
        
        with open(csv_file, 'r', encoding='utf-8') as file:
            csv_reader = csv.DictReader(file)
            
            for row_num, row in enumerate(csv_reader, start=1):
                try:
                    # Extract game data with correct column names
                    game_id = row.get('game_id', '').strip()
                    internal_name = row.get('internal_name', '').strip()
                    provider = row.get('provider.internal_name', '').strip()
                    
                    if not game_id or not internal_name:
                        print(f"⚠️  Skipping row {row_num}: Missing game_id or internal_name")
                        error_count += 1
                        continue
                    
                    # Use internal_name as the display name
                    game_name = internal_name
                    
                    # Insert or update game
                    cursor.execute("""
                        INSERT INTO games (game_id, name, internal_name, provider, integration_partner)
                        VALUES (%s, %s, %s, %s, %s)
                        ON CONFLICT (game_id) DO UPDATE SET
                            name = EXCLUDED.name,
                            internal_name = EXCLUDED.internal_name,
                            provider = EXCLUDED.provider,
                            integration_partner = EXCLUDED.integration_partner
                    """, (game_id, game_name, internal_name, provider, 'groovetech'))
                    
                    if cursor.rowcount == 1:
                        imported_count += 1
                    else:
                        updated_count += 1
                    
                    # Commit every 100 rows to avoid large transactions
                    if row_num % 100 == 0:
                        conn.commit()
                        print(f"📊 Processed {row_num} rows...")
                
                except Exception as e:
                    print(f"⚠️  Error processing row {row_num}: {str(e)}")
                    error_count += 1
                    # Rollback this specific row and continue
                    conn.rollback()
                    continue
        
        # Final commit
        conn.commit()
        
        print(f"\n✅ Import completed successfully!")
        print(f"📊 Statistics:")
        print(f"   - New games imported: {imported_count}")
        print(f"   - Existing games updated: {updated_count}")
        print(f"   - Errors encountered: {error_count}")
        print(f"   - Total processed: {imported_count + updated_count}")
        
        # Verification
        cursor.execute("SELECT COUNT(*) FROM games WHERE integration_partner = 'groovetech'")
        groovetech_count = cursor.fetchone()[0]
        
        cursor.execute("SELECT COUNT(*) FROM games")
        total_count = cursor.fetchone()[0]
        
        print(f"\n🔍 Verification Results:")
        print(f"   - Total games in database: {total_count}")
        print(f"   - GrooveTech games: {groovetech_count}")
        
        # Check specific games
        cursor.execute("SELECT game_id, name FROM games WHERE game_id IN ('158000658', '82695')")
        specific_games = cursor.fetchall()
        
        for game_id, name in specific_games:
            print(f"   - {name} ({game_id}): ✅ Found")
        
        cursor.close()
        conn.close()
        
        print(f"\n🎉 Import process completed!")
        
    except Exception as e:
        print(f"❌ Database connection failed: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    import_games()
