const { getDefaultConfig, mergeConfig } = require('@react-native/metro-config');

/**
 * Metro configuration
 * https://reactnative.dev/docs/metro
 *
 * @type {import('@react-native/metro-config').MetroConfig}
 */
const config = {
    resolver: {
        // Enable Node.js subpath package exports support for @bufbuild/protobuf
        unstable_enablePackageExports: true,
    },
};

module.exports = mergeConfig(getDefaultConfig(__dirname), config);
