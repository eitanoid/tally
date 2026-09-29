// vi: set ts=2 sw=2

export interface FormattedTimestamp {
  relative: string;
  formatted: string;
}

export function formatCreatedAt(
  createdAt?: any,
  fallbackData?: string | Record<string, any>
): FormattedTimestamp {
  let ms: number | null = null;

  if (createdAt !== null && createdAt !== undefined) {
    // A. Plain number (Unix timestamp in milliseconds or seconds)
    if (typeof createdAt === 'number') {
      if (createdAt > 1e11) {
        ms = createdAt;
      } else if (createdAt > 0) {
        ms = createdAt * 1000;
      }
    }
    // B. BigInt (Unix timestamp in seconds or milliseconds)
    else if (typeof createdAt === 'bigint') {
      const num = Number(createdAt);
      if (num > 1e11) {
        ms = num;
      } else if (num > 0) {
        ms = num * 1000;
      }
    }
    // C. Protobuf Timestamp object or WKT representation ({ seconds, nanos } or toDate() method)
    else if (typeof createdAt === 'object') {
      if (typeof createdAt.toDate === 'function') {
        try {
          const d = createdAt.toDate();
          if (d instanceof Date && !isNaN(d.getTime())) {
            ms = d.getTime();
          }
        } catch { }
      }

      if (ms === null && ('seconds' in createdAt || 'seconds_' in createdAt)) {
        const rawSec = createdAt.seconds ?? createdAt.seconds_;
        const rawNanos = createdAt.nanos ?? createdAt.nanos_ ?? 0;
        const sec = typeof rawSec === 'bigint' ? Number(rawSec) : Number(rawSec);
        if (!isNaN(sec) && sec > 0) {
          ms = sec * 1000 + Math.floor(Number(rawNanos) / 1e6);
        }
      }
    }
    // D. ISO string / RFC3339 / SQLite datetime string (e.g. "2026-09-29 18:00:00" or "2026-09-29T18:00:00Z")
    else if (typeof createdAt === 'string' && createdAt.trim().length > 0) {
      const cleanStr = createdAt.trim();
      // If sqlite format without 'T' or timezone (YYYY-MM-DD HH:MM:SS), normalize to ISO
      const normalizedStr = /^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}/.test(cleanStr)
        ? `${cleanStr.replace(' ', 'T')}Z`
        : cleanStr;
      const parsed = Date.parse(normalizedStr);
      if (!isNaN(parsed) && parsed > 0) {
        ms = parsed;
      }
    }
  }

  // Fallback: check inside fallbackData JSON for any timestamp/date fields
  if (ms === null && fallbackData) {
    try {
      const dataObj =
        typeof fallbackData === 'string' ? JSON.parse(fallbackData) : fallbackData;
      const possibleDateFields = [
        'created_at',
        'createdAt',
        'timestamp',
        'date',
        'recorded_at',
        'date_time',
      ];
      for (const f of possibleDateFields) {
        if (dataObj[f]) {
          const raw = String(dataObj[f]);
          const normalized = /^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}/.test(raw)
            ? `${raw.replace(' ', 'T')}Z`
            : raw;
          const parsed = Date.parse(normalized);
          if (!isNaN(parsed) && parsed > 0) {
            ms = parsed;
            break;
          }
        }
      }
    } catch { }
  }

  if (ms === null || isNaN(ms) || ms <= 0) {
    return { relative: 'Recent', formatted: 'Recent' };
  }

  const date = new Date(ms);

  const formatted = date.toLocaleString([], {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  const now = Date.now();
  const diffSec = Math.floor((now - ms) / 1000);

  let relative = formatted;
  if (diffSec >= 0 && diffSec < 60) {
    relative = 'Just now';
  } else if (diffSec >= 60 && diffSec < 3600) {
    const mins = Math.floor(diffSec / 60);
    relative = `${mins}m ago`;
  } else if (diffSec >= 3600 && diffSec < 86400) {
    const hours = Math.floor(diffSec / 3600);
    relative = `${hours}h ago`;
  } else if (diffSec >= 86400 && diffSec < 604800) {
    const days = Math.floor(diffSec / 86400);
    relative = `${days}d ago`;
  }

  return { relative, formatted };
}
