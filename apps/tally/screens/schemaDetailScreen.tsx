// vi: set ts=2 sw=2
import { useState, useEffect, useCallback } from 'react';
import { StyleSheet, View, ScrollView, ActivityIndicator } from 'react-native';
import {
  Text,
  IconButton,
  SegmentedButtons,
  Surface,
  Menu,
  DataTable,
  FAB,
  Button,
} from 'react-native-paper';
import { Schema, Entry } from '../generated/tally/v1/service_pb';
import { listEntries, recordEntry } from '../modules/tally-backend';
import { DynamicEntryFormModal } from '../components/EntryRecordModal';

interface SchemaDetailScreenProps {
  schema: Schema;
  onBack: () => void;
}

export function SchemaDetailScreen({ schema, onBack }: SchemaDetailScreenProps) {
  const [viewMode, setViewMode] = useState<'cards' | 'table'>('cards');
  const [entries, setEntries] = useState<Entry[]>([]);
  const [loading, setLoading] = useState(true);
  const [menuVisibleId, setMenuVisibleId] = useState<string | null>(null);

  // Form Modal state
  const [formModalVisible, setFormModalVisible] = useState(false);
  const [selectedEntry, setSelectedEntry] = useState<Entry | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const fetchEntries = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listEntries({
        tallyId: schema.tallyId,
        limit: 100,
        offset: 0,
      });
      setEntries(res.entries || []);
    } catch (err) {
      console.error('Failed to load entries:', err);
    } finally {
      setLoading(false);
    }
  }, [schema.tallyId]);

  useEffect(() => {
    fetchEntries();
  }, [fetchEntries]);

  // Parse JSON Schema properties for display and table headers
  let schemaProps: string[] = [];
  try {
    const parsed = JSON.parse(schema.jsonSchema);
    if (parsed.properties) {
      schemaProps = Object.keys(parsed.properties);
    }
  } catch (e) {
    console.error('Error parsing schema JSON:', e);
  }

  const handleRecordOrUpdate = async (formData: Record<string, any>) => {
    setIsSubmitting(true);
    try {
      await recordEntry({
        tallyId: schema.tallyId,
        schemaVersion: schema.schemaVersion,
        payloadJson: JSON.stringify(formData),
      });

      setFormModalVisible(false);
      setSelectedEntry(null);
      await fetchEntries();
    } catch (err) {
      console.error('Failed to submit entry:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <View style={styles.container}>
      {/* Header bar */}
      <View style={styles.appBar}>
        <IconButton icon="arrow-left" onPress={onBack} />
        <View style={{ flex: 1 }}>
          <Text variant="titleMedium" numberOfLines={1} style={{ fontWeight: '700' }}>
            {schema.name}
          </Text>
          <Text variant="bodySmall" numberOfLines={1} style={{ opacity: 0.6 }}>
            {schema.description}
          </Text>
        </View>
        <SegmentedButtons
          value={viewMode}
          onValueChange={(val) => setViewMode(val as 'cards' | 'table')}
          density="high"
          style={{ width: 120 }}
          buttons={[
            { value: 'cards', icon: 'view-grid-outline' },
            { value: 'table', icon: 'table' },
          ]}
        />
      </View>

      {/* Main Content */}
      {loading ? (
        <View style={styles.centered}>
          <ActivityIndicator size="large" />
        </View>
      ) : entries.length === 0 ? (
        <View style={styles.centered}>
          <Text variant="titleMedium" style={{ opacity: 0.6 }}>
            No entries logged yet
          </Text>
          <Button
            mode="outlined"
            style={{ marginTop: 12 }}
            onPress={() => {
              setSelectedEntry(null);
              setFormModalVisible(true);
            }}
          >
            Record First Entry
          </Button>
        </View>
      ) : viewMode === 'cards' ? (
        <ScrollView contentContainerStyle={styles.scrollContent}>
          {entries.map((entry) => {
            let dataObj: Record<string, any> = {};
            try {
              dataObj = JSON.parse(entry.data);
            } catch { }

            return (
              <Surface key={entry.entryId} style={styles.entryCard} elevation={1}>
                <View style={styles.entryHeader}>
                  <Text variant="labelSmall" style={{ opacity: 0.5 }}>
                    ID: {entry.entryId.slice(0, 8)}...
                  </Text>
                  <Menu
                    visible={menuVisibleId === entry.entryId}
                    onDismiss={() => setMenuVisibleId(null)}
                    anchor={
                      <IconButton
                        icon="dots-vertical"
                        size={18}
                        onPress={() => setMenuVisibleId(entry.entryId)}
                      />
                    }
                  >
                    <Menu.Item
                      onPress={() => {
                        setMenuVisibleId(null);
                        setSelectedEntry(entry);
                        setFormModalVisible(true);
                      }}
                      title="Edit"
                      leadingIcon="pencil"
                    />
                    <Menu.Item
                      onPress={() => {
                        setMenuVisibleId(null);
                        // Trigger delete call when ready
                      }}
                      title="Delete"
                      leadingIcon="delete"
                    />
                  </Menu>
                </View>

                {Object.entries(dataObj).map(([key, val]) => (
                  <View key={key} style={styles.dataRow}>
                    <Text variant="bodyMedium" style={styles.dataKey}>
                      {key}:
                    </Text>
                    <Text variant="bodyMedium" style={styles.dataVal}>
                      {String(val)}
                    </Text>
                  </View>
                ))}
              </Surface>
            );
          })}
        </ScrollView>
      ) : (
        /* Table View */
        <ScrollView horizontal style={styles.tableWrapper}>
          <ScrollView contentContainerStyle={{ paddingBottom: 80 }}>
            <DataTable>
              <DataTable.Header>
                <DataTable.Title style={{ width: 100 }}>ID</DataTable.Title>
                {schemaProps.map((prop) => (
                  <DataTable.Title key={prop} style={{ width: 120 }}>
                    {prop}
                  </DataTable.Title>
                ))}
                <DataTable.Title style={{ width: 60 }}>Actions</DataTable.Title>
              </DataTable.Header>

              {entries.map((entry) => {
                let dataObj: Record<string, any> = {};
                try {
                  dataObj = JSON.parse(entry.data);
                } catch { }

                return (
                  <DataTable.Row key={entry.entryId}>
                    <DataTable.Cell style={{ width: 100 }}>
                      {entry.entryId.slice(0, 6)}
                    </DataTable.Cell>
                    {schemaProps.map((prop) => (
                      <DataTable.Cell key={prop} style={{ width: 120 }}>
                        {dataObj[prop] !== undefined ? String(dataObj[prop]) : '-'}
                      </DataTable.Cell>
                    ))}
                    <DataTable.Cell style={{ width: 60 }}>
                      <Menu
                        visible={menuVisibleId === entry.entryId}
                        onDismiss={() => setMenuVisibleId(null)}
                        anchor={
                          <IconButton
                            icon="dots-vertical"
                            size={16}
                            onPress={() => setMenuVisibleId(entry.entryId)}
                          />
                        }
                      >
                        <Menu.Item
                          onPress={() => {
                            setMenuVisibleId(null);
                            setSelectedEntry(entry);
                            setFormModalVisible(true);
                          }}
                          title="Edit"
                        />
                        <Menu.Item
                          onPress={() => setMenuVisibleId(null)}
                          title="Delete"
                        />
                      </Menu>
                    </DataTable.Cell>
                  </DataTable.Row>
                );
              })}
            </DataTable>
          </ScrollView>
        </ScrollView>
      )}

      {/* FAB for new Entry */}
      <FAB
        icon="plus"
        style={styles.fab}
        onPress={() => {
          setSelectedEntry(null);
          setFormModalVisible(true);
        }}
      />

      {/* Dynamic Entry Recording Form */}
      <DynamicEntryFormModal
        visible={formModalVisible}
        title={selectedEntry ? 'Edit Entry' : 'Record New Entry'}
        jsonSchemaRaw={schema.jsonSchema}
        initialData={selectedEntry ? JSON.parse(selectedEntry.data || '{}') : {}}
        onDismiss={() => {
          setFormModalVisible(false);
          setSelectedEntry(null);
        }}
        onSubmit={handleRecordOrUpdate}
        isSubmitting={isSubmitting}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  appBar: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingRight: 12,
    height: 56,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  scrollContent: {
    padding: 12,
    paddingBottom: 80,
  },
  entryCard: {
    padding: 12,
    borderRadius: 8,
    marginBottom: 8,
  },
  entryHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: -8,
  },
  dataRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 2,
  },
  dataKey: {
    fontWeight: '600',
    marginRight: 6,
  },
  dataVal: {
    opacity: 0.8,
  },
  tableWrapper: {
    flex: 1,
  },
  fab: {
    position: 'absolute',
    right: 16,
    bottom: 16,
  },
});
