import React from 'react';
import { TouchableOpacity, Text, StyleSheet, ViewStyle } from 'react-native';
import { COLORS, RADIUS, SHADOWS, FONT_SIZES } from '../styles/colors';
import type { StroopColor } from '../types/game';

interface ColorButtonProps {
  color: StroopColor;
  onPress: (color: StroopColor) => void;
  disabled?: boolean;
}

export default function ColorButton({ color, onPress, disabled = false }: ColorButtonProps) {
  const getColorStyle = (): ViewStyle => {
    switch (color) {
      case 'red':
        return { backgroundColor: COLORS.red };
      case 'blue':
        return { backgroundColor: COLORS.blue };
      case 'green':
        return { backgroundColor: COLORS.green };
      case 'yellow':
        return { backgroundColor: COLORS.yellow };
    }
  };

  const getTextColor = () => {
    // Yellow needs dark text for contrast
    return color === 'yellow' ? COLORS.textDark : COLORS.white;
  };

  const handlePress = () => {
    if (!disabled) {
      onPress(color);
    }
  };

  return (
    <TouchableOpacity
      style={[
        styles.button,
        getColorStyle(),
        disabled && styles.disabled,
        !disabled && SHADOWS.medium,
      ]}
      onPress={handlePress}
      disabled={disabled}
      activeOpacity={0.7}
    >
      <Text style={[styles.text, { color: getTextColor() }]}>
        {color.toUpperCase()}
      </Text>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  button: {
    flex: 1,
    minHeight: 100,
    justifyContent: 'center',
    alignItems: 'center',
    borderRadius: RADIUS.large,
    margin: 8,
  },
  text: {
    fontSize: FONT_SIZES.xl,
    fontWeight: 'bold',
    letterSpacing: 1,
  },
  disabled: {
    opacity: 0.5,
  },
});