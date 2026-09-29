// vi: set ts=2 sw=2
import React, { useState, createContext, useContext } from 'react';
import { StyleSheet, View, ScrollView, useColorScheme } from 'react-native';
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
} from 'react-native-paper';
import './generated/tally/v1/service_pb.ts';

// --- Types ---
type ThemeMode = 'light' | 'dark' | 'auto';

interface Tally {
  id: string;
  name: string;
  description: string;
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
  const systemColorScheme = useColorScheme();
  const [themeMode, setThemeMode] = useState<ThemeMode>('auto');

  // Determine active theme based on user selection or system preference
  const isDark =
    themeMode === 'dark' || (themeMode === 'auto' && systemColorScheme === 'dark');
  const activeTheme = isDark ? darkTheme : lightTheme;

  return (
    <ThemeContext.Provider value={{ themeMode, setThemeMode }}>
      <SafeAreaProvider>
        <PaperProvider theme={activeTheme}>
          <StatusBar
            barStyle={isDark ? 'light-content' : 'dark-content'}
          />
          <MainAppContent />
        </PaperProvider>
      </SafeAreaProvider>
    </ThemeContext.Provider>
  );
}

function MainAppContent() {
  // 1. ALL HOOKS DECLARED UNCONDITIONALLY AT THE VERY TOP
  const insets = useSafeAreaInsets();
  const { themeMode, setThemeMode } = useContext(ThemeContext);

  const [tallies, setTallies] = useState<Tally[]>([
    {
      id: '1',
      name: 'Water 2',
      description: 'Daily glass count',
      count: 4,
      lastRecordedAt: '10 mins ago',
    },
    {
      id: '2',
      name: 'Workouts',
      description: 'Gym sessions this month',
      count: 12,
      lastRecordedAt: 'Yesterday',
    },
  ]);

  const [fabOpen, setFabOpen] = useState(false);
  const [createDialogVisible, setCreateDialogVisible] = useState(false);
  const [settingsDialogVisible, setSettingsDialogVisible] = useState(false);
  const [themeMenuVisible, setThemeMenuVisible] = useState(false);

  const [newTallyName, setNewTallyName] = useState('');
  const [newTallyDesc, setNewTallyDesc] = useState('');

  // 2. HANDLERS AND HELPERS
  const handleCreateTally = () => {
    if (newTallyName.trim()) {
      const newEntry: Tally = {
        id: Date.now().toString(),
        name: newTallyName.trim(),
        description: newTallyDesc.trim() || 'No description',
        count: 0,
        lastRecordedAt: 'Never',
      };
      setTallies([newEntry, ...tallies]);
      setNewTallyName('');
      setNewTallyDesc('');
      setCreateDialogVisible(false);
    }
  };

  const handleIncrement = (id: string) => {
    const nowStr = new Date().toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
    setTallies(
      tallies.map((t) =>
        t.id === id ? { ...t, count: t.count + 1, lastRecordedAt: nowStr } : t
      )
    );
  };

  const getThemeIcon = () => {
    if (themeMode === 'light') return 'weather-sunny';
    if (themeMode === 'dark') return 'weather-night';
    return 'theme-light-dark';
  };

  // 3. JSX RENDER
  return (
    <View
      style={[
        styles.container,
        { paddingTop: insets.top, paddingBottom: insets.bottom },
      ]}
    >
      {/* Compact Header */}
      <View style={styles.compactHeader}>
        <Text style={styles.headerTitle}>Tally</Text>
        <View style={styles.headerActions}>
          {/* Color Scheme Switcher */}
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

          {/* Settings Button */}
          <IconButton
            icon="cog-outline"
            size={20}
            onPress={() => setSettingsDialogVisible(true)}
          />
        </View>
      </View>

      {/* Main Content Area */}
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {tallies.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Text variant="titleMedium" style={styles.emptyText}>
              You have no tallies right now
            </Text>
            <Text variant="bodySmall" style={styles.emptySubtext}>
              Tap the + button to create your first tally.
            </Text>
          </View>
        ) : (
          tallies.map((tally) => (
            <Surface key={tally.id} style={styles.horizontalCard} elevation={1}>
              <TouchableRipple
                style={styles.cardRipple}
                onPress={() => handleIncrement(tally.id)}
                rippleColor="rgba(0, 0, 0, .1)"
              >
                <View style={styles.cardRow}>
                  {/* Left Column: Name & Description */}
                  <View style={styles.leftCol}>
                    <Text
                      variant="titleMedium"
                      style={styles.tallyName}
                      numberOfLines={1}
                    >
                      {tally.name}
                    </Text>
                    <Text
                      variant="bodySmall"
                      style={styles.tallyDescription}
                      numberOfLines={1}
                    >
                      {tally.description}
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
                if (tallies.length > 0) handleIncrement(tallies[0].id);
              },
            },
          ]}
          onStateChange={({ open }) => setFabOpen(open)}
        />

        {/* Create Tally Dialog */}
        <Dialog
          visible={createDialogVisible}
          onDismiss={() => setCreateDialogVisible(false)}
        >
          <Dialog.Title>Create New Tally</Dialog.Title>
          <Dialog.Content>
            <TextInput
              label="Tally Name"
              value={newTallyName}
              onChangeText={setNewTallyName}
              mode="outlined"
              style={{ marginBottom: 12 }}
              autoFocus
            />
            <TextInput
              label="Description (optional)"
              value={newTallyDesc}
              onChangeText={setNewTallyDesc}
              mode="outlined"
            />
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setCreateDialogVisible(false)}>
              Cancel
            </Button>
            <Button
              mode="contained"
              onPress={handleCreateTally}
              disabled={!newTallyName.trim()}
            >
              Create
            </Button>
          </Dialog.Actions>
        </Dialog>

        {/* Settings Dialog */}
        <Dialog
          visible={settingsDialogVisible}
          onDismiss={() => setSettingsDialogVisible(false)}
        >
          <Dialog.Title>Settings</Dialog.Title>
          <Dialog.Content>
            <Text variant="bodyMedium">
              Engine Version: Mock Memory Backend
            </Text>
            <Text variant="bodySmall" style={{ marginTop: 8, opacity: 0.6 }}>
              App version 0.0.1 (React Native + Paper)
            </Text>
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setSettingsDialogVisible(false)}>
              Close
            </Button>
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
