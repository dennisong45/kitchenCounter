import { format, createLogger, transports } from 'winston';
import { logs } from '@opentelemetry/api-logs';
import { LoggerProvider } from '@opentelemetry/sdk-logs';
import { HttpLogExporter } from '@opentelemetry/exporter-http-logs';

// Step 1: Create and configure the logger provider and exporter
const logExporter = new HttpLogExporter({
    url: 'http://your-log-server.com/logs', // Replace with your log server URL
});
const loggerProvider = new LoggerProvider();
loggerProvider.addLogRecordProcessor(new SimpleLogRecordProcessor(logExporter));
logs.setGlobalLoggerProvider(loggerProvider);

// Step 2: Create your custom Winston logger
export const create = function (options, args = {}) {
    const defaults = {
        console: { 
            silent: false,
            level: 'verbose',
            stderr: 'levels'
        }
    };

    // Merge defaults with options
    const config = {
        console: { ...defaults.console, ...options.console }
    };

    // Create the logger based on finalOptions
    const winstonLogger = createLogger({
        format: format.combine(
            format.timestamp(),
            format.json()
        ),
        transports: [
            new transports.Console({
                level: config.console.level,
                silent: config.console.silent,
                stderrLevels: config.console.stderr
            })
        ]
    });

    return winstonLogger;
};

// Usage example
const logger = create({
    console: { 
        level: 'info',
        silent: false
    }
});

// Step 3: Integrate with OpenTelemetry logging
const otelLogger = logs.getLogger('my-logger', '1.0.0');

// Custom method to log and send logs to remote server using OpenTelemetry
function logWithOtel(severity, message) {
    logger.log(severity, message); // Log locally using Winston
    otelLogger.emit({ severityNumber: severityMap[severity], body: message }); // Send to remote server
}

// Map Winston log levels to OpenTelemetry severity numbers
const severityMap = {
    error: 17, // Example mapping, adjust as needed
    warn: 13,
    info: 9,
    debug: 5,
    trace: 1
};

// Example usage
logWithOtel('info', 'This is an info message');
logWithOtel('error', 'This is an error message');
