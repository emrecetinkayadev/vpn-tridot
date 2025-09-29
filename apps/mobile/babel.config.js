const path = require('path');

module.exports = {
  presets: ['module:@react-native/babel-preset'],
  plugins: [
    [
      'module-resolver',
      {
        alias: {
          '@vpn/mobile-core': path.resolve(__dirname, '../packages/mobile-core/src'),
        },
        extensions: ['.ios.ts', '.android.ts', '.ts', '.tsx', '.js', '.json'],
      },
    ],
  ],
};
