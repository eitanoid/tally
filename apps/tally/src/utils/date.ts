import { format, isValid, parseISO } from 'date-fns';

export interface FormattedTimestamp {
    relative: string;
    formatted: string;
}

/**
 * Safely converts an ISO string or Date into a formatted string.
 */
export function formatDate(
    dateInput: string | Date | null | undefined,
    formatPattern: string = 'PPP' // Default: e.g. "Sep 26, 2026"
): string {
    if (!dateInput) return '';

    const date = typeof dateInput === 'string' ? parseISO(dateInput) : dateInput;

    if (!isValid(date)) return '';

    return format(date, formatPattern);
}

/**
 * Format ISO 8601 timestamps (e.g. "2026-09-26T15:00:00Z") for entry tables and cards
 */
export function formatDateTime(isoString: string): string {
    return formatDate(isoString, 'MMM d, yyyy • h:mm a'); // e.g. "Sep 26, 2026 • 3:00 PM"
}

/**
 * Format ISO full dates (e.g. "2026-09-26")
 */
export function formatFullDate(dateString: string): string {
    return formatDate(dateString, 'MMM d, yyyy'); // e.g. "Sep 26, 2026"
}

/**
 * Convert user Go-duration strings ("15m30s", "28s") into human-friendly UI text
 */
export function formatGoDuration(durationStr: string): string {
    if (!durationStr) return '';

    // Go durations support decimals and ns/us/µs/ms/s/m/h units.
    const match = durationStr.match(
        /^([+-]?)(?:(?:\d+(?:\.\d*)?|\.\d+)(?:ns|us|µs|μs|ms|s|m|h))+$/
    );
    if (!match) return durationStr;

    const [, sign] = match;
    const body = durationStr.slice(sign.length).replace(/μs/g, 'µs');
    const parts = body.match(/(?:\d+(?:\.\d*)?|\.\d+)(?:ns|us|µs|ms|s|m|h)/g);

    return parts ? `${sign}${parts.join(' ')}` : durationStr;
}

/**
 * Format a protobuf timestamp (or a date found in entry data) for entry cards and tables.
 */
export function formatCreatedAt(
    createdAt?: unknown,
    fallbackData?: string | Record<string, unknown>
): FormattedTimestamp {
    let date = toDate(createdAt);

    if (!date && fallbackData) {
        try {
            const data = typeof fallbackData === 'string' ? JSON.parse(fallbackData) : fallbackData;
            const dateKeys = ['created_at', 'createdAt', 'timestamp', 'date', 'recorded_at', 'date_time'];
            for (const key of dateKeys) {
                date = toDate(data[key]);
                if (date) break;
            }
        } catch {
            // Ignore malformed entry data and fall back to the generic label.
        }
    }

    if (!date) return { relative: 'Recent', formatted: 'Recent' };

    const formatted = format(date, 'MMM d, yyyy, HH:mm');
    const elapsedSeconds = Math.floor((Date.now() - date.getTime()) / 1000);
    let relative = formatted;

    if (elapsedSeconds >= 0 && elapsedSeconds < 60) {
        relative = 'Just now';
    } else if (elapsedSeconds >= 60 && elapsedSeconds < 3600) {
        relative = `${Math.floor(elapsedSeconds / 60)}m ago`;
    } else if (elapsedSeconds >= 3600 && elapsedSeconds < 86400) {
        relative = `${Math.floor(elapsedSeconds / 3600)}h ago`;
    } else if (elapsedSeconds >= 86400 && elapsedSeconds < 604800) {
        relative = `${Math.floor(elapsedSeconds / 86400)}d ago`;
    }

    return { relative, formatted };
}

function toDate(value: unknown): Date | null {
    if (value instanceof Date) return isValid(value) ? value : null;

    if (typeof value === 'number' || typeof value === 'bigint') {
        const numericValue = Number(value);
        if (!Number.isFinite(numericValue)) return null;
        const milliseconds = Math.abs(numericValue) > 1e11 ? numericValue : numericValue * 1000;
        const date = new Date(milliseconds);
        return isValid(date) ? date : null;
    }

    if (typeof value === 'string' && value.trim()) {
        const normalized = value.trim().replace(
            /^(\d{4}-\d{2}-\d{2})\s(\d{2}:\d{2}:\d{2})$/,
            '$1T$2Z'
        );
        const date = parseISO(normalized);
        return isValid(date) ? date : null;
    }

    if (value && typeof value === 'object') {
        const timestamp = value as {
            seconds?: number | string | bigint;
            seconds_?: number | string | bigint;
            nanos?: number;
            nanos_?: number;
            toDate?: () => Date;
        };

        if (typeof timestamp.toDate === 'function') {
            try {
                const date = timestamp.toDate();
                if (date instanceof Date && isValid(date)) return date;
            } catch {
                // Fall through to the seconds/nanos representation.
            }
        }

        const rawSeconds = timestamp.seconds ?? timestamp.seconds_;
        if (rawSeconds !== undefined) {
            const seconds = Number(rawSeconds);
            const nanos = Number(timestamp.nanos ?? timestamp.nanos_ ?? 0);
            if (Number.isFinite(seconds) && Number.isFinite(nanos)) {
                const date = new Date(seconds * 1000 + Math.floor(nanos / 1e6));
                return isValid(date) ? date : null;
            }
        }
    }

    return null;
}
