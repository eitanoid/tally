// vi: set ts=2 sw=2
import { useState } from 'react';
import { StyleSheet, ScrollView } from 'react-native';
import { DataTable, IconButton, Menu } from 'react-native-paper';
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
  const [activeMenuId, setActiveMenuId] = useState<string | null>(null);

  return (
    <ScrollView horizontal style={styles.tableWrapper}>
      <ScrollView contentContainerStyle={styles.scrollContent}>
        <DataTable>
          <DataTable.Header>
            <DataTable.Title style={styles.dateCol}>Recorded At</DataTable.Title>
            {schemaProps.map((prop) => (
              <DataTable.Title key={prop} style={styles.dataCol}>
                {prop}
              </DataTable.Title>
            ))}
            <DataTable.Title style={styles.actionsCol}>Actions</DataTable.Title>
          </DataTable.Header>

          {entries.map((entry) => {
            let dataObj: Record<string, any> = {};
            try {
              dataObj = JSON.parse(entry.data || '{}');
            } catch { }

            const timeInfo = formatCreatedAt(entry.createdAt, entry.data);

            return (
              <DataTable.Row key={entry.entryId}>
                <DataTable.Cell style={styles.dateCol}>
                  {timeInfo.formatted}
                </DataTable.Cell>
                {schemaProps.map((prop) => (
                  <DataTable.Cell key={prop} style={styles.dataCol}>
                    {dataObj[prop] !== undefined ? String(dataObj[prop]) : '-'}
                  </DataTable.Cell>
                ))}
                <DataTable.Cell style={styles.actionsCol}>
                  {(onEdit || onDelete) && (
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
                  )}
                </DataTable.Cell>
              </DataTable.Row>
            );
          })}
        </DataTable>
      </ScrollView>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  tableWrapper: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 80,
  },
  dateCol: {
    width: 150,
  },
  dataCol: {
    width: 120,
  },
  actionsCol: {
    width: 60,
  },
});
