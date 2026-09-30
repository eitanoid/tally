// vi: set ts=2 sw=2
import React from 'react';
import { StyleSheet, View } from 'react-native';
import { Dialog, Portal, Button, Text, useTheme } from 'react-native-paper';

export interface DeleteConfirmationDialogProps {
  visible: boolean;
  title?: string;
  message?: string;
  itemName?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  isDeleting?: boolean;
  onConfirm: () => void | Promise<void>;
  onDismiss: () => void;
}

export function DeleteConfirmationDialog({
  visible,
  title = 'Delete Confirmation',
  message = 'Are you sure you want to delete this? This action cannot be undone.',
  itemName,
  confirmLabel = 'Delete',
  cancelLabel = 'Cancel',
  isDeleting = false,
  onConfirm,
  onDismiss,
}: DeleteConfirmationDialogProps) {
  const theme = useTheme();

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={onDismiss} style={styles.dialog}>
        <Dialog.Title style={styles.title}>{title}</Dialog.Title>
        <Dialog.Content>
          {itemName ? (
            <Text variant="bodyLarge" style={styles.itemName}>
              "{itemName}"
            </Text>
          ) : null}
          <Text
            variant="bodyMedium"
            style={[styles.message, { color: theme.colors.onSurfaceVariant }]}
          >
            {message}
          </Text>
        </Dialog.Content>
        <Dialog.Actions style={styles.actions}>
          <Button onPress={onDismiss} disabled={isDeleting}>
            {cancelLabel}
          </Button>
          <Button
            mode="contained"
            buttonColor={theme.colors.error}
            textColor={theme.colors.onError}
            loading={isDeleting}
            disabled={isDeleting}
            onPress={onConfirm}
          >
            {confirmLabel}
          </Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
}

// Re-export with modal alias for naming convenience
export { DeleteConfirmationDialog as DeleteConfirmationModal };

const styles = StyleSheet.create({
  dialog: {
    borderRadius: 16,
  },
  title: {
    fontWeight: '600',
  },
  itemName: {
    fontWeight: 'bold',
    marginBottom: 8,
  },
  message: {
    lineHeight: 20,
  },
  actions: {
    paddingHorizontal: 16,
    paddingBottom: 8,
  },
});
