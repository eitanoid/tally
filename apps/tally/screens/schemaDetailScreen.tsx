// vi: set ts=2 sw=2
import { useState, useEffect, useCallback } from 'react';
import { StyleSheet, View, ScrollView, ActivityIndicator, BackHandler, Alert } from 'react-native';
import { Text, FAB, Button, useTheme } from 'react-native-paper';
import { Schema, Entry, ResponseCode } from '../generated/tally/v1/service_pb';
import { listEntries, recordEntry, updateEntry, deleteEntry, deleteTally } from '../modules/tally-backend';
import { DynamicEntryFormModal } from '../components/DynamicEntryFormModal';
import { DeleteConfirmationDialog } from '../components/DeleteConfirmationDialog';
import { SchemaDetailHeader } from '../components/SchemaDetailHeader';
import { EntryCard } from '../components/EntryCard';
import { EntryTable } from '../components/EntryTable';

interface SchemaDetailScreenProps {
  schema: Schema;
  onBack: () => void;
}

export function SchemaDetailScreen({ schema, onBack }: SchemaDetailScreenProps) {
  const theme = useTheme();
  const [viewMode, setViewMode] = useState<'cards' | 'table'>('cards');
  const [entries, setEntries] = useState<Entry[]>([]);
  const [loading, setLoading] = useState(true);

  // Form Modal state
  const [formModalVisible, setFormModalVisible] = useState(false);
  const [selectedEntry, setSelectedEntry] = useState<Entry | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Delete Confirmation state
  const [entryToDelete, setEntryToDelete] = useState<Entry | null>(null);
  const [deleteSchemaDialogOpen, setDeleteSchemaDialogOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  // Intercept device back button to return to tally list instead of exiting app
  useEffect(() => {
    const onBackPress = () => {
      if (entryToDelete) {
        setEntryToDelete(null);
        return true;
      }
      if (deleteSchemaDialogOpen) {
        setDeleteSchemaDialogOpen(false);
        return true;
      }
      if (formModalVisible) {
        setFormModalVisible(false);
        setSelectedEntry(null);
        return true;
      }
      onBack();
      return true;
    };
    const sub = BackHandler.addEventListener('hardwareBackPress', onBackPress);
    return () => sub.remove();
  }, [entryToDelete, deleteSchemaDialogOpen, formModalVisible, onBack]);

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
      if (selectedEntry) {
        const res = await updateEntry({
          entryId: selectedEntry.entryId,
          patchData: JSON.stringify(formData),
        });
        if (res.code !== ResponseCode.OK) {
          throw new Error(res.errorMessage || 'Failed to update entry');
        }
      } else {
        const res = await recordEntry({
          tallyId: schema.tallyId,
          schemaVersion: schema.schemaVersion,
          payloadJson: JSON.stringify(formData),
        });
        if (res.code !== ResponseCode.OK) {
          throw new Error(res.errorMessage || 'Failed to record entry');
        }
      }

      setFormModalVisible(false);
      setSelectedEntry(null);
      await fetchEntries();
    } catch (err: any) {
      console.error('Failed to submit entry:', err);
      Alert.alert('Error', err.message || 'Failed to submit entry');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEditEntry = (entry: Entry) => {
    setSelectedEntry(entry);
    setFormModalVisible(true);
  };

  const handleDeleteEntry = (entry: Entry) => {
    setEntryToDelete(entry);
  };

  const confirmDeleteEntry = async () => {
    if (!entryToDelete) return;
    setIsDeleting(true);
    try {
      const res = await deleteEntry({
        entryId: entryToDelete.entryId,
      });
      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to delete entry');
      }
      setEntryToDelete(null);
      await fetchEntries();
    } catch (error: any) {
      console.error('Failed to delete entry:', error);
      Alert.alert('Error', error.message || 'Failed to delete entry');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleExportData = () => {
    // Export data formatting
    console.log('Export data requested for schema:', schema.name, entries);
  };

  const handleDeleteSchema = () => {
    setDeleteSchemaDialogOpen(true);
  };

  const confirmDeleteSchema = async () => {
    setIsDeleting(true);
    try {
      const res = await deleteTally({
        tallyId: schema.tallyId,
      });
      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to delete tally');
      }
      setDeleteSchemaDialogOpen(false);
      onBack();
    } catch (error: any) {
      console.error('Failed to delete schema:', error);
      Alert.alert('Error', error.message || 'Failed to delete schema');
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <View style={[styles.container, { backgroundColor: theme.colors.background }]}>
      {/* Header Bar and Metadata Summary */}
      <SchemaDetailHeader
        schema={schema}
        totalEntries={entries.length}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        onBack={onBack}
        onExportData={handleExportData}
        onDeleteSchema={handleDeleteSchema}
      />

      {/* Main Content */}
      {loading ? (
        <View style={styles.centered}>
          <ActivityIndicator size="large" />
        </View>
      ) : entries.length === 0 ? (
        <View style={styles.centered}>
          <Text variant="titleMedium" style={styles.emptyTitle}>
            No entries logged yet
          </Text>
          <Button
            mode="outlined"
            style={styles.emptyButton}
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
          {entries.map((entry) => (
            <EntryCard
              key={entry.entryId}
              entry={entry}
              onEdit={handleEditEntry}
              onDelete={handleDeleteEntry}
            />
          ))}
        </ScrollView>
      ) : (
        <EntryTable
          entries={entries}
          schemaProps={schemaProps}
          onEdit={handleEditEntry}
          onDelete={handleDeleteEntry}
        />
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

      {/* Dynamic Entry Recording Form with locked schema */}
      <DynamicEntryFormModal
        visible={formModalVisible}
        title={selectedEntry ? 'Edit Entry' : 'Record New Entry'}
        schema={schema}
        lockSchema={true}
        initialData={selectedEntry ? JSON.parse(selectedEntry.data || '{}') : {}}
        onDismiss={() => {
          setFormModalVisible(false);
          setSelectedEntry(null);
        }}
        onSubmit={handleRecordOrUpdate}
        isSubmitting={isSubmitting}
      />

      {/* Delete Entry Confirmation Dialog */}
      <DeleteConfirmationDialog
        visible={Boolean(entryToDelete)}
        title="Delete Entry"
        message="Are you sure you want to delete this recorded entry? This cannot be undone."
        isDeleting={isDeleting}
        onConfirm={confirmDeleteEntry}
        onDismiss={() => setEntryToDelete(null)}
      />

      {/* Delete Schema Confirmation Dialog */}
      <DeleteConfirmationDialog
        visible={deleteSchemaDialogOpen}
        title="Delete Tally Schema"
        itemName={schema.name}
        message="Are you sure you want to delete this tally? All recorded entries associated with it will also be permanently deleted."
        confirmLabel="Delete Tally"
        isDeleting={isDeleting}
        onConfirm={confirmDeleteSchema}
        onDismiss={() => setDeleteSchemaDialogOpen(false)}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  emptyTitle: {
    opacity: 0.6,
  },
  emptyButton: {
    marginTop: 12,
  },
  scrollContent: {
    padding: 12,
    paddingBottom: 80,
  },
  fab: {
    position: 'absolute',
    right: 16,
    bottom: 16,
  },
});
