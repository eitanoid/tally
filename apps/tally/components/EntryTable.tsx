// vi: set ts=2 sw=2
import { useState } from 'react';
import { StyleSheet, ScrollView, View } from 'react-native';
import { DataTable, IconButton, Menu, useTheme } from 'react-native-paper';
import { Entry } from '../generated/tally/v1/service_pb';
import { formatCreatedAt } from '../src/utils/date';

interface EntryTableProps {
  entries: Entry[];
  schemaProps: string[];
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
}

export function EntryTable({
  entries,
  schemaProps,
  onEdit,
  onDelete,
}: EntryTableProps) {
  const theme = useTheme();
  const [activeMenuId, setActiveMenuId] = useState<string | null>(null);

  const isWide = schemaProps.length > 2;

  const content = (
    <DataTable style={[styles.dataTable, !isWide && styles.fullWidthTable]}>
      <DataTable.Header style={{ borderBottomColor: theme.colors.outlineVariant }}>
        <DataTable.Title
          style={isWide ? styles.dateColFixed : styles.dateColFlex}
          textStyle={styles.headerText}
        >
          Recorded At
        </DataTable.Title>
        {schemaProps.map((prop) => (
          <DataTable.Title
            key={prop}
            style={isWide ? styles.dataColFixed : styles.dataColFlex}
            textStyle={styles.headerText}
          >
            {prop}
          </DataTable.Title>
        ))}
        {(onEdit || onDelete) && (
          <DataTable.Title
            style={styles.actionsCol}
            numeric
          >
            Actions
          </DataTable.Title>
        )}
      </DataTable.Header>

      {entries.map((entry) => {
        let dataObj: Record<string, any> = {};
        try {
          dataObj = JSON.parse(entry.data || '{}');
        } catch { }

        const timeInfo = formatCreatedAt(entry.createdAt, entry.data);

        return (
          <DataTable.Row
            key={entry.entryId}
            style={{ borderBottomColor: theme.colors.surfaceVariant }}
          >
            <DataTable.Cell
              style={isWide ? styles.dateColFixed : styles.dateColFlex}
            >
              {timeInfo.formatted}
            </DataTable.Cell>
            {schemaProps.map((prop) => (
              <DataTable.Cell
                key={prop}
                style={isWide ? styles.dataColFixed : styles.dataColFlex}
              >
                {dataObj[prop] !== undefined ? String(dataObj[prop]) : '-'}
              </DataTable.Cell>
            ))}
            {(onEdit || onDelete) && (
              <DataTable.Cell style={styles.actionsCol} numeric>
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

  if (isWide) {
    return (
      <ScrollView
        horizontal
        style={styles.horizontalScroll}
        contentContainerStyle={styles.horizontalContent}
        showsHorizontalScrollIndicator={true}
      >
        <ScrollView
          style={styles.verticalScroll}
          contentContainerStyle={styles.verticalContent}
        >
          {content}
        </ScrollView>
      </ScrollView>
    );
  }

  return (
    <ScrollView
      style={styles.verticalScroll}
      contentContainerStyle={styles.verticalContent}
    >
      <View style={styles.fullWidthContainer}>
        {content}
      </View>
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
  fullWidthContainer: {
    width: '100%',
    paddingHorizontal: 8,
  },
  dataTable: {
    minWidth: '100%',
  },
  fullWidthTable: {
    width: '100%',
  },
  headerText: {
    fontWeight: '700',
  },
  dateColFixed: {
    width: 150,
  },
  dataColFixed: {
    width: 130,
  },
  dateColFlex: {
    flex: 1.4,
  },
  dataColFlex: {
    flex: 1,
  },
  actionsCol: {
    width: 60,
    justifyContent: 'center',
    alignItems: 'center',
  },
});
