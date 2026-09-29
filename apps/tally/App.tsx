// vi: set ts=2 sw=2
import React, { useState, useEffect, createContext, useContext, useCallback } from 'react';
import { StyleSheet, View, ScrollView, useColorScheme, ActivityIndicator } from 'react-native';
import { StatusBar } from 'react-native';
import {
  SafeAreaProvider,
  useSafeAreaInsets,
} from 'react-native-safe-area-context';
import {
  PaperProvider,
  MD3LightTheme,
  MD3DarkTheme,
  Surface,
  Text,
  FAB,
  Portal,
  Dialog,
  TextInput,
  Button,
  TouchableRipple,
  IconButton,
  Menu,
  SegmentedButtons,
} from 'react-native-paper';

import {
  pingGo,
  listSchemas,
  createSchema,
  listEntries,
  recordEntry,
} from './modules/tally-backend';
import {
  FieldFormat,
  ResponseCode,
  type Schema,
} from './generated/tally/v1/service_pb';

// --- Types ---
type ThemeMode = 'light' | 'dark' | 'auto';

interface TallyViewModel {
  schema: Schema;
  count: number;
  lastRecordedAt: string;
}

// --- Themes ---
const lightTheme = {
  ...MD3LightTheme,
  colors: {
    ...MD3LightTheme.colors,
    primary: '#1e88e5',
    secondary: '#00acc1',
  },
};

const darkTheme = {
  ...MD3DarkTheme,
  colors: {
    ...MD3DarkTheme.colors,
    primary: '#90caf9',
    secondary: '#80deea',
  },
};

// --- Theme Context ---
const ThemeContext = createContext<{
  themeMode: ThemeMode;
  setThemeMode: (mode: ThemeMode) => void;
}>({
  themeMode: 'auto',
  setThemeMode: () => { },
});

export default function App(): React.JSX.Element {
  console.log('Go Engine Ping:', pingGo('Eitan'));

  const systemColorScheme = useColorScheme();
  const [themeMode, setThemeMode] = useState<ThemeMode>('auto');

  const isDark =
    themeMode === 'dark' || (themeMode === 'auto' && systemColorScheme === 'dark');
  const activeTheme = isDark ? darkTheme : lightTheme;

  return (
    <ThemeContext.Provider value={{ themeMode, setThemeMode }}>
      <SafeAreaProvider>
        <PaperProvider theme={activeTheme}>
          <StatusBar barStyle={isDark ? 'light-content' : 'dark-content'} />
          <MainAppContent />
        </PaperProvider>
      </SafeAreaProvider>
    </ThemeContext.Provider>
  );
}

function MainAppContent() {
  const insets = useSafeAreaInsets();
  const { themeMode, setThemeMode } = useContext(ThemeContext);

  // --- Backend State ---
  const [tallies, setTallies] = useState<TallyViewModel[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // --- UI Dialog & Menu States ---
  const [fabOpen, setFabOpen] = useState(false);
  const [createDialogVisible, setCreateDialogVisible] = useState(false);
  const [settingsDialogVisible, setSettingsDialogVisible] = useState(false);
  const [themeMenuVisible, setThemeMenuVisible] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // --- Create Tally Form State ---
  const [newTallyName, setNewTallyName] = useState('');
  const [newTallyDesc, setNewTallyDesc] = useState('');
  const [newFieldName, setNewFieldName] = useState('count');
  const [selectedFieldFormat, setSelectedFieldFormat] = useState<FieldFormat>(
    FieldFormat.INTEGER
  );

  // --- 1. FETCH ALL SCHEMAS & ENTRY STATS FROM GO BACKEND ---
  const fetchTalliesFromBackend = useCallback(async () => {
    setLoading(true);
    setErrorMsg(null);
    try {
      // Call listSchemas FFI RPC
      const res = await listSchemas({});
      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to list schemas');
      }

      // Fetch count and last recorded entry for each schema in parallel
      const loadedTallies: TallyViewModel[] = await Promise.all(
        res.schemas.map(async (schema) => {
          try {
            const entriesRes = await listEntries({
              tallyId: schema.tallyId,
              limit: 1,
              offset: 0,
            });

            const count = entriesRes.totalCount ?? 0;
            let lastRecordedAt = 'Never';

            if (entriesRes.entries.length > 0 && entriesRes.entries[0].createdAt) {
              const seconds = Number(entriesRes.entries[0].createdAt.seconds);
              if (!isNaN(seconds) && seconds > 0) {
                lastRecordedAt = new Date(seconds * 1000).toLocaleTimeString([], {
                  hour: '2-digit',
                  minute: '2-digit',
                });
              }
            }

            return { schema, count, lastRecordedAt };
          } catch {
            return { schema, count: 0, lastRecordedAt: 'Error loading' };
          }
        })
      );

      setTallies(loadedTallies);
    } catch (err: any) {
      console.error('Failed to load tallies:', err);
      setErrorMsg(err.message || 'Error connecting to Go backend');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTalliesFromBackend();
  }, [fetchTalliesFromBackend]);

  // --- 2. CREATE SCHEMA RPC HANDLER ---
  const handleCreateTally = async () => {
    if (!newTallyName.trim()) return;

    setIsSubmitting(true);
    try {
      const res = await createSchema({
        name: newTallyName.trim(),
        description: newTallyDesc.trim() || 'No description',
        fields: [
          {
            $typeName: 'tally.v1.SchemaRequestField',
            name: newFieldName.trim() || 'value',
            description: 'Default field value',
            type: selectedFieldFormat,
            required: true,
          },
        ],
      });

      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to create schema');
      }

      // Reset Form State
      setNewTallyName('');
      setNewTallyDesc('');
      setNewFieldName('count');
      setCreateDialogVisible(false);

      // Refresh list from Go SQLite database
      await fetchTalliesFromBackend();
    } catch (err: any) {
      console.error('Create Schema Error:', err);
      setErrorMsg(err.message || 'Failed to create schema');
    } finally {
      setIsSubmitting(false);
    }
  };

  // --- 3. RECORD ENTRY RPC HANDLER ---
  const handleIncrement = async (tally: TallyViewModel) => {
    try {
      const payload = JSON.stringify({ count: tally.count + 1, timestamp: Date.now() });

      const res = await recordEntry({
        tallyId: tally.schema.tallyId,
        schemaVersion: tally.schema.schemaVersion,
        payloadJson: payload,
      });

      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to record entry');
      }

      // Refresh view model with newly recorded entry
      await fetchTalliesFromBackend();
    } catch (err: any) {
      console.error('Record Entry Error:', err);
    }
  };

  const getThemeIcon = () => {
    if (themeMode === 'light') return 'weather-sunny';
    if (themeMode === 'dark') return 'weather-night';
    return 'theme-light-dark';
  };

  return (
    <View style={[styles.container, { paddingTop: insets.top, paddingBottom: insets.bottom }]}>
      {/* Compact Header */}
      <View style={styles.compactHeader}>
        <Text style={styles.headerTitle}>Tally</Text>
        <View style={styles.headerActions}>
          <Menu
            visible={themeMenuVisible}
            onDismiss={() => setThemeMenuVisible(false)}
            anchor={
              <IconButton
                icon={getThemeIcon()}
                size={20}
                onPress={() => setThemeMenuVisible(true)}
              />
            }
          >
            <Menu.Item
              onPress={() => {
                setThemeMode('light');
                setThemeMenuVisible(false);
              }}
              title="Light"
              leadingIcon="weather-sunny"
            />
            <Menu.Item
              onPress={() => {
                setThemeMode('dark');
                setThemeMenuVisible(false);
              }}
              title="Dark"
              leadingIcon="weather-night"
            />
            <Menu.Item
              onPress={() => {
                setThemeMode('auto');
                setThemeMenuVisible(false);
              }}
              title="System Auto"
              leadingIcon="theme-light-dark"
            />
          </Menu>

          <IconButton
            icon="cog-outline"
            size={20}
            onPress={() => setSettingsDialogVisible(true)}
          />
        </View>
      </View>

      {/* Main Content Area */}
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {loading ? (
          <View style={styles.emptyContainer}>
            <ActivityIndicator size="large" />
            <Text variant="bodyMedium" style={{ marginTop: 12 }}>
              Loading schemas from Go SQLite...
            </Text>
          </View>
        ) : errorMsg ? (
          <View style={styles.emptyContainer}>
            <Text variant="titleMedium" style={{ color: 'red', marginBottom: 8 }}>
              Backend Error
            </Text>
            <Text variant="bodySmall">{errorMsg}</Text>
            <Button mode="outlined" style={{ marginTop: 16 }} onPress={fetchTalliesFromBackend}>
              Retry
            </Button>
          </View>
        ) : tallies.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Text variant="titleMedium" style={styles.emptyText}>
              You have no tallies right now
            </Text>
            <Text variant="bodySmall" style={styles.emptySubtext}>
              Tap the + button to create your first tally schema in SQLite.
            </Text>
          </View>
        ) : (
          tallies.map((tally) => (
            <Surface key={tally.schema.tallyId} style={styles.horizontalCard} elevation={1}>
              <TouchableRipple
                style={styles.cardRipple}
                onPress={() => handleIncrement(tally)}
                rippleColor="rgba(0, 0, 0, .1)"
              >
                <View style={styles.cardRow}>
                  {/* Left Column: Name & Description */}
                  <View style={styles.leftCol}>
                    <Text variant="titleMedium" style={styles.tallyName} numberOfLines={1}>
                      {tally.schema.name}
                    </Text>
                    <Text variant="bodySmall" style={styles.tallyDescription} numberOfLines={1}>
                      {tally.schema.description}
                    </Text>
                  </View>

                  {/* Right Column: Count & Timestamp */}
                  <View style={styles.rightCol}>
                    <Text variant="headlineMedium" style={styles.tallyCount}>
                      {tally.count}
                    </Text>
                    <Text variant="labelSmall" style={styles.lastRecorded}>
                      {tally.lastRecordedAt}
                    </Text>
                  </View>
                </View>
              </TouchableRipple>
            </Surface>
          ))
        )}
      </ScrollView>

      {/* Speed Dial Expanding FAB */}
      <Portal>
        <FAB.Group
          open={fabOpen}
          visible
          icon={fabOpen ? 'close' : 'plus'}
          actions={[
            {
              icon: 'file-document-plus-outline',
              label: 'Create New Tally',
              onPress: () => setCreateDialogVisible(true),
            },
            {
              icon: 'playlist-plus',
              label: 'Record Entry',
              onPress: () => {
                if (tallies.length > 0) handleIncrement(tallies[0]);
              },
            },
          ]}
          onStateChange={({ open }) => setFabOpen(open)}
        />

        {/* Create Tally Dialog */}
        <Dialog visible={createDialogVisible} onDismiss={() => setCreateDialogVisible(false)}>
          <Dialog.Title>Create New Tally Schema</Dialog.Title>
          <Dialog.Content>
            <TextInput
              label="Tally Name"
              value={newTallyName}
              onChangeText={setNewTallyName}
              mode="outlined"
              style={{ marginBottom: 8 }}
              autoFocus
            />
            <TextInput
              label="Description (optional)"
              value={newTallyDesc}
              onChangeText={setNewTallyDesc}
              mode="outlined"
              style={{ marginBottom: 12 }}
            />
            <TextInput
              label="Default Field Name"
              value={newFieldName}
              onChangeText={setNewFieldName}
              mode="outlined"
              style={{ marginBottom: 12 }}
            />
            <Text variant="bodySmall" style={{ marginBottom: 6 }}>
              Field Data Type:
            </Text>
            <SegmentedButtons
              value={selectedFieldFormat.toString()}
              onValueChange={(val) => setSelectedFieldFormat(Number(val) as FieldFormat)}
              buttons={[
                { value: FieldFormat.INTEGER.toString(), label: 'Integer' },
                { value: FieldFormat.STRING.toString(), label: 'String' },
                { value: FieldFormat.BOOLEAN.toString(), label: 'Bool' },
              ]}
            />
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setCreateDialogVisible(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button
              mode="contained"
              onPress={handleCreateTally}
              disabled={!newTallyName.trim() || isSubmitting}
              loading={isSubmitting}
            >
              Create
            </Button>
          </Dialog.Actions>
        </Dialog>

        {/* Settings Dialog */}
        <Dialog visible={settingsDialogVisible} onDismiss={() => setSettingsDialogVisible(false)}>
          <Dialog.Title>Settings</Dialog.Title>
          <Dialog.Content>
            <Text variant="bodyMedium">Engine: Go SQLite Bridge (gomobile + Protobuf FFI)</Text>
            <Text variant="bodySmall" style={{ marginTop: 8, opacity: 0.6 }}>
              App version 0.0.1 (React Native + Paper)
            </Text>
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setSettingsDialogVisible(false)}>Close</Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  compactHeader: {
    height: 48,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingLeft: 16,
    paddingRight: 4,
  },
  headerTitle: {
    fontWeight: '800',
    fontSize: 20,
  },
  headerActions: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  scrollContent: {
    paddingHorizontal: 12,
    paddingTop: 8,
    paddingBottom: 100,
  },
  emptyContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingTop: 140,
  },
  emptyText: {
    opacity: 0.7,
    marginBottom: 4,
  },
  emptySubtext: {
    opacity: 0.5,
  },
  horizontalCard: {
    marginBottom: 8,
    borderRadius: 8,
    overflow: 'hidden',
  },
  cardRipple: {
    paddingVertical: 8,
    paddingHorizontal: 12,
  },
  cardRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  leftCol: {
    flex: 1,
    paddingRight: 12,
  },
  rightCol: {
    alignItems: 'flex-end',
  },
  tallyName: {
    fontWeight: '700',
    fontSize: 16,
    lineHeight: 20,
  },
  tallyDescription: {
    opacity: 0.6,
    fontSize: 12,
    marginTop: 2,
  },
  tallyCount: {
    fontWeight: 'bold',
    fontSize: 22,
    lineHeight: 26,
  },
  lastRecorded: {
    opacity: 0.4,
    fontSize: 10,
    marginTop: 1,
  },
});
