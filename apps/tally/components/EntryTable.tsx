// vi: set ts=2 sw=2
import { useMemo, useState } from 'react';
import { StyleSheet, ScrollView } from 'react-native';
import { DataTable, IconButton, Menu, useTheme } from 'react-native-paper';
import { Entry, FieldFormat } from '../generated/tally/v1/service_pb';
import { formatCreatedAt } from '../src/utils/date';
import { formatFieldValue } from '../src/fields/fieldFormat';

const EMPTY_FIELD_FORMATS: Record<string, FieldFormat> = {};

function getColumnWidth(values: string[], minWidth: number): number {
  const longestValueWidth = Math.max(
    0,
    ...values.map((value) =>
      Array.from(value).reduce(
        (width, character) => width + (character.codePointAt(0)! > 0x2ff ? 16 : 12),
        32
      )
    )
  );
  return Math.max(minWidth, longestValueWidth);
}

interface EntryTableProps {
  entries: Entry[];
  schemaProps: string[];
  fieldFormats?: Record<string, FieldFormat>;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
}

export function EntryTable({
  entries,
  schemaProps,
  fieldFormats = EMPTY_FIELD_FORMATS,
  onEdit,
  onDelete,
}: EntryTableProps) {
  const theme = useTheme();
  const [activeMenuId, setActiveMenuId] = useState<string | null>(null);

  const rows = useMemo(
    () =>
      entries.map((entry) => {
        let dataObj: Record<string, any> = {};
        try {
          dataObj = JSON.parse(entry.data || '{}');
        } catch { }

        return {
          entry,
          dataObj,
          timeInfo: formatCreatedAt(entry.createdAt, entry.data),
        };
      }),
    [entries]
  );

  const dateColumnWidth = getColumnWidth(
    ['Recorded At', ...rows.map((row) => row.timeInfo.formatted)],
    150
  );
  const dataColumnWidths = schemaProps.map((prop) =>
    getColumnWidth(
      [
        prop,
        ...rows.map((row) =>
          row.dataObj[prop] !== undefined
            ? formatFieldValue(row.dataObj[prop], fieldFormats[prop] ?? FieldFormat.UNSPECIFIED)
            : '-'
        ),
      ],
      130
    )
  );
  const hasActions = Boolean(onEdit || onDelete);
  const actionsColumnWidth = hasActions ? getColumnWidth(['Actions'], 60) : 0;
  const tableWidth =
    dateColumnWidth +
    dataColumnWidths.reduce((total, width) => total + width, 0) +
    actionsColumnWidth +
    32;

  const content = (
    <DataTable style={{ width: tableWidth }}>
      <DataTable.Header style={{ borderBottomColor: theme.colors.outlineVariant }}>
        <DataTable.Title
          style={[styles.fixedColumn, { width: dateColumnWidth }]}
          textStyle={styles.headerText}
        >
          Recorded At
        </DataTable.Title>
        {schemaProps.map((prop, index) => (
          <DataTable.Title
            key={prop}
            style={[styles.fixedColumn, { width: dataColumnWidths[index] }]}
            textStyle={styles.headerText}
          >
            {prop}
          </DataTable.Title>
        ))}
        {hasActions && (
          <DataTable.Title
            style={[styles.fixedColumn, { width: actionsColumnWidth }]}
            numeric
          >
            Actions
          </DataTable.Title>
        )}
      </DataTable.Header>

      {rows.map(({ entry, dataObj, timeInfo }) => {
        return (
          <DataTable.Row
            key={entry.entryId}
            style={{ borderBottomColor: theme.colors.surfaceVariant }}
          >
            <DataTable.Cell
              style={[styles.fixedColumn, { width: dateColumnWidth }]}
            >
              {timeInfo.formatted}
            </DataTable.Cell>
            {schemaProps.map((prop, index) => (
              <DataTable.Cell
                key={prop}
                style={[styles.fixedColumn, { width: dataColumnWidths[index] }]}
              >
                {dataObj[prop] !== undefined
                  ? formatFieldValue(dataObj[prop], fieldFormats[prop] ?? FieldFormat.UNSPECIFIED)
                  : '-'}
              </DataTable.Cell>
            ))}
            {hasActions && (
              <DataTable.Cell
                style={[styles.fixedColumn, { width: actionsColumnWidth }]}
                numeric
              >
                <Menu
                  visible={activeMenuId === entry.entryId}
                  onDismiss={() => setActiveMenuId(null)}
                  anchor={
                    <IconButton
                      icon="dots-vertical"
                      size={16}
                      onPress={() => setActiveMenuId(entry.entryId)}
                    />
                  }
                >
                  {onEdit && (
                    <Menu.Item
                      onPress={() => {
                        setActiveMenuId(null);
                        onEdit(entry);
                      }}
                      title="Edit"
                    />
                  )}
                  {onDelete && (
                    <Menu.Item
                      onPress={() => {
                        setActiveMenuId(null);
                        onDelete(entry);
                      }}
                      title="Delete"
                    />
                  )}
                </Menu>
              </DataTable.Cell>
            )}
          </DataTable.Row>
        );
      })}
    </DataTable>
  );

  return (
    <ScrollView
      horizontal
      style={styles.horizontalScroll}
      contentContainerStyle={styles.horizontalContent}
      showsHorizontalScrollIndicator
    >
      <ScrollView
        style={[styles.verticalScroll, { width: tableWidth }]}
        contentContainerStyle={styles.verticalContent}
      >
        {content}
      </ScrollView>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  horizontalScroll: {
    flex: 1,
    width: '100%',
  },
  horizontalContent: {
    minWidth: '100%',
  },
  verticalScroll: {
    flex: 1,
    width: '100%',
  },
  verticalContent: {
    paddingBottom: 90,
  },
  fixedColumn: {
    flex: 0,
    flexGrow: 0,
    flexShrink: 0,
  },
  headerText: {
    fontWeight: '700',
  },
});
