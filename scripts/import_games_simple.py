#!/usr/bin/env python3
"""
Simple Game Import Script - Just INSERT without ON CONFLICT
"""

import csv
import psycopg2
import sys

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
        
        # Clear existing games first
        cursor.execute("DELETE FROM games WHERE integration_partner = 'groovetech'")
        conn.commit()
        print("🗑️  Cleared existing GrooveTech games")
        
        # Read CSV file
        csv_file = 'game 20250930-905.csv'
        imported_count = 0
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
                        error_count += 1
                        continue
                    
                    # Use internal_name as the display name
                    game_name = internal_name
                    
                    # Simple INSERT
                    cursor.execute("""
                        INSERT INTO games (game_id, name, internal_name, provider, integration_partner)
                        VALUES (%s, %s, %s, %s, %s)
                    """, (game_id, game_name, internal_name, provider, 'groovetech'))
                    
                    imported_count += 1
                    
                    # Commit every 100 rows
                    if row_num % 100 == 0:
                        conn.commit()
                        print(f"📊 Processed {row_num} rows...")
                
                except Exception as e:
                    error_count += 1
                    conn.rollback()
                    continue
        
        # Final commit
        conn.commit()
        
        print(f"\n✅ Import completed successfully!")
        print(f"📊 Statistics:")
        print(f"   - New games imported: {imported_count}")
        print(f"   - Errors encountered: {error_count}")
        
        # Verification
        cursor.execute("SELECT COUNT(*) FROM games WHERE integration_partner = 'groovetech'")
        groovetech_count = cursor.fetchone()[0]
        
        cursor.execute("SELECT COUNT(*) FROM games")
        total_count = cursor.fetchone()[0]
        
        print(f"\n🔍 Verification Results:")
        print(f"   - Total games in database: {total_count}")
        print(f"   - GrooveTech games: {groovetech_count}")
        
        cursor.close()
        conn.close()
        
        print(f"\n🎉 Import process completed!")
        
    except Exception as e:
        print(f"❌ Database connection failed: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    import_games()
