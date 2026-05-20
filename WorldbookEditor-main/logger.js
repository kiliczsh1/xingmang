// logger.js - Centralized logging system for Worldbook Editor

const LOG_PREFIX = '[世界书管理器]';

class Logger {
    constructor() {
        this.logs = [];
        this.maxLogs = 500; // Keep last 500 logs
    }

    _log(level, message, ...args) {
        const timestamp = new Date().toISOString();
        const fullMessage = `${LOG_PREFIX} ${message}`;

        // Store log entry
        this.logs.push({
            timestamp,
            level,
            message: fullMessage,
            args: args.length > 0 ? args : undefined
        });

        // Trim if exceeds max
        if (this.logs.length > this.maxLogs) {
            this.logs.shift();
        }

        // Output to console with color coding
        const levelColors = {
            info: '#3b82f6',
            warn: '#f59e0b',
            error: '#ef4444',
            debug: '#6b7280'
        };
        const color = levelColors[level] || '#000000';
        console.log(`%c${timestamp} ${fullMessage}`, `color:${color}; font-weight:bold;`, ...args);
    }

    info(message, ...args) { this._log('info', message, ...args); }
    warn(message, ...args) { this._log('warn', message, ...args); }
    error(message, ...args) { this._log('error', message, ...args); }
    debug(message, ...args) { this._log('debug', message, ...args); }

    getLogs(filterLevel = null) {
        if (!filterLevel || filterLevel === 'all') {
            return this.logs;
        }
        return this.logs.filter(log => log.level === filterLevel);
    }

    clear() {
        this.logs = [];
    }

    exportText() {
        return this.logs.map(log => {
            let msg = log.message;
            if (log.args && log.args.length > 0) {
                msg += ' ' + log.args.map(arg =>
                    typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
                ).join(' ');
            }
            return `[${log.timestamp}] [${log.level.toUpperCase()}] ${msg}`;
        }).join('\n');
    }

    exportJSON() {
        return JSON.stringify(this.logs, null, 2);
    }
}

export const logger = new Logger();
