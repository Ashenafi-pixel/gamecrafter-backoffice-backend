#!/usr/bin/env python3
"""
Migration script to copy data from PostgreSQL to ClickHouse
Handles: transactions (deposits/withdrawals), groove_transactions, bets, 
cashback_earnings, cashback_claims, withdrawals, wallet_transactions
"""

import psycopg2
from clickhouse_driver import Client
from datetime import datetime
import json

# PostgreSQL connection
pg_conn = psycopg2.connect(
    host="localhost",
    port=5433,
    database="tucanbit",
    user="tucanbit",
    password="5kj0YmV5FKKpU9D50B7yH5A"
)

# ClickHouse connection
ch_client = Client(
    host='localhost',
    port=9000,
    user='tucanbit',
    password='tucanbit_clickhouse_password',
    database='tucanbit_analytics'
)

def migrate_groove_transactions():
    """Migrate groove_transactions from PostgreSQL to ClickHouse"""
    print("Migrating groove_transactions...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT id, transaction_id, account_id, session_id, amount, currency, type, status, created_at, metadata
            FROM groove_transactions
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} groove_transactions to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.groove_transactions (id, transaction_id, account_id, session_id, amount, currency, type, status, created_at, metadata) VALUES",
                [(str(row[0]), row[1], row[2], row[3] if row[3] else '', row[4], row[5], row[6], row[7], row[8], json.dumps(row[9]) if row[9] else '') for row in rows]
            )
            print(f"Migrated {len(rows)} groove_transactions")

def migrate_transactions_as_deposits():
    """Migrate deposit transactions from transactions table to deposits table in ClickHouse"""
    print("Migrating deposits from transactions table...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT id, user_id, amount, currency_code, status, created_at, updated_at
            FROM transactions
            WHERE transaction_type = 'deposit'
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} deposit transactions to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.deposits (id, user_id, amount, currency, status, created_at, updated_at) VALUES",
                [(str(row[0]), str(row[1]), row[2], row[3], row[4], row[5] or datetime.now(), row[6] or datetime.now()) for row in rows]
            )
            print(f"Migrated {len(rows)} deposits")

def migrate_transactions_as_withdrawals():
    """Migrate withdrawal transactions from transactions table to withdrawals table in ClickHouse"""
    print("Migrating withdrawals from transactions table...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT id, user_id, amount, currency_code, status, created_at, updated_at
            FROM transactions
            WHERE transaction_type = 'withdrawal'
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} withdrawal transactions to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.withdrawals (id, user_id, amount, currency, status, created_at, updated_at) VALUES",
                [(str(row[0]), str(row[1]), row[2], row[3], row[4], row[5] or datetime.now(), row[6] or datetime.now()) for row in rows]
            )
            print(f"Migrated {len(rows)} withdrawals")

def migrate_withdrawals_table():
    """Migrate withdrawals table data to ClickHouse"""
    print("Migrating withdrawals table...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT withdrawal_id, user_id, usd_amount_cents, currency_code, status, created_at, updated_at
            FROM withdrawals
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} withdrawals to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.withdrawals (id, user_id, amount, currency, status, created_at, updated_at) VALUES",
                [(str(row[0]), str(row[1]), row[2] / 100.0, row[3] if row[3] else 'USD', row[4], row[5] or datetime.now(), row[6] or datetime.now()) for row in rows]
            )
            print(f"Migrated {len(rows)} withdrawals from withdrawals table")

def migrate_bets():
    """Migrate bets from PostgreSQL to ClickHouse"""
    print("Migrating bets...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT id, user_id, amount, currency, status, timestamp
            FROM bets
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} bets to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.bets (id, user_id, amount, currency, status, created_at) VALUES",
                [(str(row[0]), str(row[1]), row[2], row[3], row[4] if row[4] else 'completed', row[5] or datetime.now()) for row in rows]
            )
            print(f"Migrated {len(rows)} bets")

def migrate_cashback_earnings():
    """Migrate cashback_earnings from PostgreSQL to ClickHouse"""
    print("Migrating cashback_earnings...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT id, user_id, tier_id, earning_type, source_bet_id, ggr_amount, cashback_rate, earned_amount, claimed_amount, available_amount, status, expires_at, claimed_at, created_at
            FROM cashback_earnings
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} cashback_earnings to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.cashback_earnings (id, user_id, tier_id, earning_type, source_bet_id, ggr_amount, cashback_rate, earned_amount, claimed_amount, available_amount, status, expires_at, claimed_at, created_at) VALUES",
                [(str(row[0]), str(row[1]), str(row[2]), row[3], str(row[4]) if row[4] else None, row[5], row[6], row[7], row[8] or 0, row[9], row[10] or 'pending', row[11] or datetime.now(), row[12], row[13] or datetime.now()) for row in rows]
            )
            print(f"Migrated {len(rows)} cashback_earnings")

def migrate_cashback_claims():
    """Migrate cashback_claims from PostgreSQL to ClickHouse"""
    print("Migrating cashback_claims...")
    
    with pg_conn.cursor() as cursor:
        cursor.execute("""
            SELECT id, user_id, claim_amount, currency_code, status, transaction_id, processing_fee, net_amount, claimed_earnings, admin_notes, processed_at, created_at
            FROM cashback_claims
        """)
        
        rows = cursor.fetchall()
        print(f"Found {len(rows)} cashback_claims to migrate")
        
        if rows:
            ch_client.execute(
                "INSERT INTO tucanbit_analytics.cashback_claims (id, user_id, claim_amount, currency_code, status, transaction_id, processing_fee, net_amount, claimed_earnings, admin_notes, processed_at, created_at) VALUES",
                [(str(row[0]), str(row[1]), row[2], row[3] if row[3] else 'USD', row[4] or 'pending', str(row[5]) if row[5] else None, row[6] or 0, row[7], json.dumps(row[8]) if row[8] else '{}', row[9], row[10], row[11] or datetime.now()) for row in rows]
            )
            print(f"Migrated {len(rows)} cashback_claims")

if __name__ == "__main__":
    try:
        print("Starting migration from PostgreSQL to ClickHouse...")
        
        # Migrate all tables
        migrate_groove_transactions()
        migrate_transactions_as_deposits()
        migrate_transactions_as_withdrawals()
        migrate_withdrawals_table()
        migrate_bets()
        migrate_cashback_earnings()
        migrate_cashback_claims()
        
        print("\nMigration completed successfully!")
        
        # Verify counts
        print("\nVerifying data in ClickHouse:")
        result = ch_client.execute("SELECT COUNT(*) FROM tucanbit_analytics.groove_transactions")
        print(f"  groove_transactions: {result[0][0]} rows")
        
        result = ch_client.execute("SELECT COUNT(*) FROM tucanbit_analytics.deposits")
        print(f"  deposits: {result[0][0]} rows")
        
        result = ch_client.execute("SELECT COUNT(*) FROM tucanbit_analytics.withdrawals")
        print(f"  withdrawals: {result[0][0]} rows")
        
        result = ch_client.execute("SELECT COUNT(*) FROM tucanbit_analytics.bets")
        print(f"  bets: {result[0][0]} rows")
        
        result = ch_client.execute("SELECT COUNT(*) FROM tucanbit_analytics.cashback_earnings")
        print(f"  cashback_earnings: {result[0][0]} rows")
        
        result = ch_client.execute("SELECT COUNT(*) FROM tucanbit_analytics.cashback_claims")
        print(f"  cashback_claims: {result[0][0]} rows")
        
    except Exception as e:
        print(f"Error during migration: {e}")
        import traceback
        traceback.print_exc()
    finally:
        pg_conn.close()
        ch_client.disconnect()
