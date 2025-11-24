export const COLORS = {
  // Primary gradient colors
  primary: '#667eea',
  primaryDark: '#764ba2',
  
  // Stroop game colors (must match backend and COLOR_MAP in game.ts)
  red: '#FF0000',
  blue: '#0000FF',
  green: '#00FF00',
  yellow: '#FFD700',
  
  // UI colors
  background: '#f8f9fa',
  white: '#ffffff',
  
  // Text colors
  textDark: '#333333',
  textMedium: '#666666',
  textLight: '#999999',
  
  // Status colors
  success: '#43e97b',
  error: '#ff6b6b',
  warning: '#ffd93d',
  
  // Border
  border: '#e0e0e0',
};

// Gradient definitions for React Native LinearGradient
export const GRADIENTS = {
  primary: ['#667eea', '#764ba2'],
  red: ['#ff6b6b', '#ee5a6f'],
  blue: ['#4facfe', '#00f2fe'],
  green: ['#43e97b', '#38f9d7'],
  yellow: ['#ffd93d', '#fcbf49'],
};

// Shadow styles for React Native
export const SHADOWS = {
  small: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2, // Android
  },
  medium: {
    shadowColor: '#667eea',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  large: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.3,
    shadowRadius: 16,
    elevation: 10,
  },
};

// Border radius constants
export const RADIUS = {
  small: 8,
  medium: 12,
  large: 16,
  xlarge: 20,
  round: 9999,
};

// Spacing constants
export const SPACING = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
  xxl: 48,
};

// Font sizes
export const FONT_SIZES = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 18,
  xl: 24,
  xxl: 32,
  xxxl: 48,
};