#!/usr/bin/env python3
"""
Import GrooveTech games from CSV into the games table
"""

import csv
import psycopg2
import sys
from datetime import datetime

def import_games():
    # Database connection (Docker container)
    conn = psycopg2.connect(
        host="localhost",
        port="5433",  # Docker port mapping
        database="tucanbit",
        user="tucanbit",
        password="5kj0YmV5FKKpU9D50B7yH5A"
    )
    
    cursor = conn.cursor()
    
    try:
        # Read CSV file
        csv_file = "game 20250930-905.csv"
        
        with open(csv_file, 'r', encoding='utf-8') as file:
            csv_reader = csv.DictReader(file)
            
            imported_count = 0
            skipped_count = 0
            
            for row in csv_reader:
                game_id = row['game_id'].strip('"')
                internal_name = row['internal_name'].strip('"')
                provider = row['provider.internal_name'].strip('"')
                
                # Check if game already exists
                cursor.execute("SELECT id FROM games WHERE game_id = %s", (game_id,))
                if cursor.fetchone():
                    print(f"Game {game_id} already exists, skipping...")
                    skipped_count += 1
                    continue
                
                # Insert new game
                cursor.execute("""
                    INSERT INTO games (name, game_id, internal_name, integration_partner, provider, status, enabled, timestamp)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                """, (
                    internal_name,  # name
                    game_id,        # game_id
                    internal_name,  # internal_name
                    'groovetech',   # integration_partner
                    provider,       # provider
                    'ACTIVE',       # status
                    True,           # enabled
                    datetime.now()  # timestamp
                ))
                
                imported_count += 1
                
                if imported_count % 100 == 0:
                    print(f"Imported {imported_count} games...")
            
            # Commit all changes
            conn.commit()
            
            print(f"\nImport completed!")
            print(f"Imported: {imported_count} games")
            print(f"Skipped: {skipped_count} games (already exist)")
            
    except Exception as e:
        print(f"Error: {e}")
        conn.rollback()
        sys.exit(1)
    
    finally:
        cursor.close()
        conn.close()

if __name__ == "__main__":
    import_games()
