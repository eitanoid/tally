// vi: set ts=2 sw=2

export interface FormattedTimestamp {
  relative: string;
  formatted: string;
}

export function formatCreatedAt(
  createdAt?: { seconds?: bigint | number | string; nanos?: number } | string | null,
  fallbackData?: string | Record<string, any>
): FormattedTimestamp {
  let ms: number | null = null;

  // 1. Try protobuf Timestamp
  if (
    createdAt &&
    typeof createdAt === 'object' &&
    'seconds' in createdAt &&
    createdAt.seconds !== undefined &&
    createdAt.seconds !== null
  ) {
    const sec =
      typeof createdAt.seconds === 'bigint'
        ? Number(createdAt.seconds)
        : Number(createdAt.seconds);
    if (!isNaN(sec) && sec > 0) {
      ms = sec * 1000 + (createdAt.nanos ? Math.floor(Number(createdAt.nanos) / 1e6) : 0);
    }
  } else if (typeof createdAt === 'string' && createdAt.trim().length > 0) {
    const parsed = Date.parse(createdAt);
    if (!isNaN(parsed) && parsed > 0) {
      ms = parsed;
    }
  }

  // 2. Fallback: check fields in entry.data (e.g. date-time, date, or timestamp fields)
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
          const parsed = Date.parse(String(dataObj[f]));
          if (!isNaN(parsed) && parsed > 0) {
            ms = parsed;
            break;
          }
        }
      }
    } catch { }
  }

  if (ms === null || isNaN(ms) || ms <= 0) {
    return { relative: 'Recent', formatted: 'Recorded' };
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
