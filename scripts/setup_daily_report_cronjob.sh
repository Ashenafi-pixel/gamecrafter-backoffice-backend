#!/bin/bash

# TucanBIT Daily Report Email Cronjob Setup Script
# This script sets up automatic daily report email sending

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration variables
SCRIPT_DIR="/path/to/tucanbit-backend"
RECIPIENTS="ashenafialemu66@gmail.com"  # Add more emails separated by commas if needed
TIME_OF_DAY="09"  # Send at 9:00 AM
TIMEZONE="UTC"    # Change to your preferred timezone
LOG_FILE="/var/log/tucanbit_daily_reports.log"

echo -e "${BLUE} Setting up TucanBIT Daily Report Email Cronjob${NC}"
echo "======================================================"

# Check if running as root or if user has sudo privileges
if ! (( $EUID == 0 )) && ! groups $USER | grep -q sudo; then
    echo -e "${RED}❌ This script requires sudo privileges to install cron job${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl is required but not installed${NC}"
    echo "Installing curl..."
    sudo apt-get update && sudo apt-get install -y curl
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq is recommended for JSON parsing${NC}"
    echo "Installing jq..."
    sudo apt-get install -y jq
fi

# Function to get application status
check_app_status() {
    local app_url="http://localhost:8080/analytics/realtime/stats"
    local response=$(curl -s -w "%{http_code}" -o /dev/null "$app_url" 2>/dev/null || echo "000")
    
    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✅ TucanBIT application is running${NC}"
        return 0
    else
        echo -e "${RED}❌ TucanBIT application is not accessible at $app_url${NC}"
        echo "   Returned HTTP status: $response"
        return 1
    fi
}

# Function to send test daily report
send_test_report() {
    echo -e "${BLUE}📧 Sending test daily report email...${NC}"
    
    # Get yesterday's date
    local yesterday=$(date -d "yesterday" "+%Y-%m-%d")
    
    # Send test request
    local response=$(curl -s -X POST "http://localhost:8080/analytics/daily-report/send" \
        -H "Content-Type: application/json" \
        -d '{
            "date": "'$yesterday'",
            "recipients": ["'$RECIPIENTS'"]
        }' 2>/dev/null)
    
    if echo "$response" | grep -q '"success":true'; then
        echo -e "${GREEN}✅ Test daily report sent successfully${NC}"
        echo "   Check your email inbox for the report!"
    else
        echo -e "${RED}❌ Failed to send test daily report${NC}"
        echo "   Response: $response"
    fi
}

# Function to create the cron script
create_cron_script() {
    echo -e "${BLUE}📄 Creating cron execution script...${NC}"
    
    local cron_script="/usr/local/bin/tucanbit-daily-report"
    
    sudo tee "$cron_script" > /dev/null <<EOF
#!/bin/bash

# TucanBIT Daily Report Email Script
# Generated on $(date)

# Configuration
APP_URL="http://localhost:8080"
RECIPIENTS="$RECIPIENTS"
LOG_FILE="$LOG_FILE"

# Function to log messages
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] \$1" >> "\$LOG_FILE"
}

# Get yesterday's date
REPORT_DATE=\$(date -d "yesterday" "+%Y-%m-%d")

log_message "Starting daily report generation for date: \$REPORT_DATE"

# Check if TucanBIT application is running
APP_STATUS=\$(curl -s -w "%{http_code}" -o /dev/null "\$APP_URL/analytics/realtime/stats" 2>/dev/null || echo "000")

if [ "\$APP_STATUS" != "200" ]; then
    log_message "ERROR: TucanBIT application not accessible (HTTP \$APP_STATUS)"
    log_message "Daily report email NOT sent for \$REPORT_DATE"
    exit 1
fi

# Send daily report email
RESPONSE=\$(curl -s -X POST "\$APP_URL/analytics/daily-report/send" \\
    -H "Content-Type: application/json" \\
    -d '{
        "date": "'\$REPORT_DATE'",
        "recipients": ["'$RECIPIENTS'"]
    }' 2>/dev/null)

if echo "\$RESPONSE" | grep -q '"success":true'; then
    log_message "SUCCESS: Daily report email sent for \$REPORT_DATE"
    log_message "Recipients: $RECIPIENTS"
else
    log_message "ERROR: Failed to send daily report email for \$REPORT_DATE"
    log_message "Response: \$RESPONSE"
    exit 1
fi

log_message "Daily report cron job completed successfully"
EOF

    sudo chmod +x "$cron_script"
    echo -e "${GREEN}✅ Cron script created at: $cron_script${NC}"
}

# Function to setup cron job
setup_cron_job() {
    echo -e "${BLUE}⏰ Setting up cron job...${NC}"
    
    local cron_entry="$TIME_OF_DAY $TIMEZONE * * * $USER /usr/local/bin/tucanbit-daily-report"
    local cron_file="/tmp/tucanbit_cron_$USER.txt"
    
    # Backup existing crontab
    crontab -l 2>/dev/null > "$cron_file" || touch "$cron_file"
    
    # Remove any existing TucanBIT daily report entries
    sed -i '/tucanbit-daily-report/d' "$cron_file"
    
    # Add new cron entry
    echo "$cron_entry" >> "$cron_file"
    
    # Install the crontab
    crontab "$cron_file"
    
    # Clean up temporary file
    rm "$cron_file"
    
    echo -e "${GREEN}✅ Cron job installed successfully${NC}"
    echo -e "${BLUE}   Schedule: Every day at $TIME_OF_DAY:00 $TIMEZONE${NC}"
    echo -e "${BLUE}   Recipients: $RECIPIENTS${NC}"
    echo -e "${BLUE}   Log file: $LOG_FILE${NC}"
}

# Function to verify cron job
verify_cron_job() {
    echo -e "${BLUE}🔍 Verifying cron job installation...${NC}"
    
    if crontab -l 2>/dev/null | grep -q "tucanbit-daily-report"; then
        echo -e "${GREEN}✅ Cron job properly installed${NC}"
        echo -e "${BLUE}Cron jobs for user $USER:${NC}"
        crontab -l 2>/dev/null | grep -v '^#' | grep -E "(tucanbit|daily-report)" || echo "No relevant cron jobs found"
    else
        echo -e "${RED}❌ Cron job installation verification failed${NC}"
        echo -e "${YELLOW}Current crontab for user $USER:${NC}"
        crontab -l 2>/dev/null || echo "No crontab found"
    fi
}

# Function to create log rotation
setup_log_rotation() {
    echo -e "${BLUE}📋 Setting up log rotation...${NC}"
    
    sudo tee /etc/logrotate.d/tucanbit-daily-reports > /dev/null <<EOF
$LOG_FILE {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 $USER $USER
    su $USER $USER
}
EOF

    echo -e "${GREEN}✅ Log rotation configured${NC}"
    echo -e "${BLUE}   Logs will be rotated daily and kept for 30 days${NC}"
}

# Main execution
echo -e "${BLUE}🔧 Starting Daily Report Email Setup...${NC}"

# Check application status
if ! check_app_status; then
    echo -e "${RED}❌ Cannot proceed - TucanBIT application is not running${NC}"
    echo -e "${YELLOW}Please ensure the application is running on port 8080 before continuing${NC}"
    exit 1
fi

# Create cron script
create_cron_script

# Setup log rotation
setup_log_rotation

# Setup cron job
setup_cron_job

# Verify installation
verify_cron_job

# Send test report (optional)
echo ""
read -p "Would you like to send a test daily report email? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    send_test_report
fi

echo ""
echo -e "${GREEN}🎉 TucanBIT Daily Report Email Setup Completed!${NC}"
echo "======================================================"
echo -e "${BLUE}📋 Summary:${NC}"
echo -e "   • API Endpoint: http://localhost:8080/analytics/daily-report/send"
echo -e "   • Schedule: Every day at $TIME_OF_DAY:00 $TIMEZONE"
echo -e "   • Recipients: $RECIPIENTS"
echo -e "   • Cron Script: /usr/local/bin/tucanbit-daily-report"
echo -e "   • Log File: $LOG_FILE"
echo ""
echo -e "${BLUE}📧 Manual Testing Commands:${NC}"
echo "   # Send yesterday's report:"
echo "   curl -X POST http://localhost:8080/analytics/daily-report/yesterday \\"
echo "        -H 'Content-Type: application/json' \\"
echo "        -d '{\"recipients\": [\"$RECIPIENTS\"]}'"
echo ""
echo "   # Send report for specific date:"
echo "   curl -X POST http://localhost:8080/analytics/daily-report/send \\"
echo "        -H 'Content-Type: application/json' \\"
echo "        -d '{\"date\": \"2025-01-15\", \"recipients\": [\"$RECIPIENTS\"]}'"
echo ""
echo -e "${BLUE}📊 Monitoring Commands:${NC}"
echo "   # View cron logs:"
echo "   tail -f $LOG_FILE"
echo ""
echo "   # Check current crontab:"
echo "   crontab -l"
echo ""
echo -e "${GREEN}✅ Setup complete! Daily reports will be sent automatically starting tomorrow.${NC}"