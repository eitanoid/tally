// vi: set ts=2 sw=2
import React, { useState, useEffect, createContext, useContext, useCallback } from 'react';
import { StyleSheet, View, ScrollView, useColorScheme, ActivityIndicator } from 'react-native';
import { StatusBar } from 'react-native';
import { SafeAreaProvider, useSafeAreaInsets } from 'react-native-safe-area-context';
import {
  PaperProvider,
  MD3LightTheme,
  MD3DarkTheme,
  Text,
  FAB,
  Portal,
  Dialog,
  Button,
  IconButton,
  Menu,
} from 'react-native-paper';

import {
  pingGo,
  listSchemas,
  createSchema,
  listEntries,
} from './modules/tally-backend';
import {
  ResponseCode,
  type Schema,
} from './generated/tally/v1/service_pb';

import { SchemaCard, TallyViewModel } from './components/SchemaCard';
import { SchemaCreateModal, DynamicField } from './components/SchemaCreateModal';
import { SchemaDetailScreen } from './screens/schemaDetailScreen';

type ThemeMode = 'light' | 'dark' | 'auto';

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
          <MainAppRouter />
        </PaperProvider>
      </SafeAreaProvider>
    </ThemeContext.Provider>
  );
}

function MainAppRouter() {
  const insets = useSafeAreaInsets();
  const { themeMode, setThemeMode } = useContext(ThemeContext);

  // --- Router / Screen state ---
  const [selectedSchema, setSelectedSchema] = useState<Schema | null>(null);

  // --- Backend State ---
  const [tallies, setTallies] = useState<TallyViewModel[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // --- UI Dialog & Menu States ---
  const [createDialogVisible, setCreateDialogVisible] = useState(false);
  const [settingsDialogVisible, setSettingsDialogVisible] = useState(false);
  const [themeMenuVisible, setThemeMenuVisible] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const fetchTalliesFromBackend = useCallback(async () => {
    setLoading(true);
    setErrorMsg(null);
    try {
      const res = await listSchemas({});
      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to list schemas');
      }

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

  const handleCreateTally = async (
    name: string,
    description: string,
    fields: DynamicField[]
  ) => {
    setIsSubmitting(true);
    try {
      const res = await createSchema({
        name,
        description: description || 'No description',
        fields: fields.map((f) => ({
          $typeName: 'tally.v1.SchemaRequestField',
          name: f.name,
          description: f.description,
          type: f.format,
          required: f.required,
        })),
      });

      if (res.code !== ResponseCode.OK) {
        throw new Error(res.errorMessage || 'Failed to create schema');
      }

      setCreateDialogVisible(false);
      await fetchTalliesFromBackend();
    } catch (err: any) {
      console.error('Create Schema Error:', err);
      setErrorMsg(err.message || 'Failed to create schema');
    } finally {
      setIsSubmitting(false);
    }
  };

  const getThemeIcon = () => {
    if (themeMode === 'light') return 'weather-sunny';
    if (themeMode === 'dark') return 'weather-night';
    return 'theme-light-dark';
  };

  // Render Detail Screen if a schema is selected
  if (selectedSchema) {
    return (
      <View style={[styles.container, { paddingTop: insets.top, paddingBottom: insets.bottom }]}>
        <SchemaDetailScreen
          schema={selectedSchema}
          onBack={() => {
            setSelectedSchema(null);
            fetchTalliesFromBackend();
          }}
        />
      </View>
    );
  }

  return (
    <View style={[styles.container, { paddingTop: insets.top, paddingBottom: insets.bottom }]}>
      {/* App Bar Header */}
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

      {/* Main List */}
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
            <Button
              mode="outlined"
              style={{ marginTop: 16 }}
              onPress={fetchTalliesFromBackend}
            >
              Retry
            </Button>
          </View>
        ) : tallies.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Text variant="titleMedium" style={styles.emptyText}>
              You have no tallies right now
            </Text>
            <Text variant="bodySmall" style={styles.emptySubtext}>
              Tap the + button to create your first tally schema.
            </Text>
          </View>
        ) : (
          tallies.map((tally) => (
            <SchemaCard
              key={tally.schema.tallyId}
              tally={tally}
              onPress={() => setSelectedSchema(tally.schema)}
            />
          ))
        )}
      </ScrollView>

      {/* Create Tally FAB */}
      <FAB
        icon="plus"
        style={styles.fab}
        onPress={() => setCreateDialogVisible(true)}
      />

      {/* Schema Creation Modal */}
      <SchemaCreateModal
        visible={createDialogVisible}
        onDismiss={() => setCreateDialogVisible(false)}
        onSubmit={handleCreateTally}
        isSubmitting={isSubmitting}
      />

      {/* Settings Dialog */}
      <Portal>
        <Dialog
          visible={settingsDialogVisible}
          onDismiss={() => setSettingsDialogVisible(false)}
        >
          <Dialog.Title>Settings</Dialog.Title>
          <Dialog.Content>
            <Text variant="bodyMedium">
              Engine: Go SQLite Bridge (gomobile + Protobuf FFI)
            </Text>
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
  fab: {
    position: 'absolute',
    right: 16,
    bottom: 16,
  },
});
