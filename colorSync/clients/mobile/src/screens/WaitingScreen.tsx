import React, { useEffect, useRef } from 'react';
import { View, Text, StyleSheet, Animated } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { GameFlowState } from '../types';

interface WaitingScreenProps {
  flowState: GameFlowState;
  username: string;
  roomId: string;
}

export default function WaitingScreen({ flowState, username, roomId }: WaitingScreenProps) {
  const spinValue = useRef(new Animated.Value(0)).current;
  const pulseValue = useRef(new Animated.Value(1)).current;

  useEffect(() => {
    // Spinning animation
    Animated.loop(
      Animated.timing(spinValue, {
        toValue: 1,
        duration: 2000,
        useNativeDriver: true,
      })
    ).start();

    // Pulsing animation
    Animated.loop(
      Animated.sequence([
        Animated.timing(pulseValue, {
          toValue: 1.2,
          duration: 800,
          useNativeDriver: true,
        }),
        Animated.timing(pulseValue, {
          toValue: 1,
          duration: 800,
          useNativeDriver: true,
        }),
      ])
    ).start();
  }, [spinValue, pulseValue]);

  const spin = spinValue.interpolate({
    inputRange: [0, 1],
    outputRange: ['0deg', '360deg'],
  });

  const getStatusMessage = () => {
    switch (flowState) {
      case 'JOINING':
        return 'Joining matchmaking...';
      case 'WAITING_ROOM':
        return 'Waiting for opponent...';
      case 'CONNECTING':
        return 'Connecting to game...';
      default:
        return 'Loading...';
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.emoji}>🎨</Text>
          <Text style={styles.title}>COLOR SYNC</Text>
          <Text style={styles.username}>Playing as: {username}</Text>
        </View>

        {/* Loading Spinner */}
        <Animated.View
          style={[
            styles.spinner,
            {
              transform: [{ rotate: spin }, { scale: pulseValue }],
            },
          ]}
        >
          <Text style={styles.spinnerText}>⚡</Text>
        </Animated.View>

        {/* Status */}
        <View style={styles.statusContainer}>
          <Text style={styles.statusText}>{getStatusMessage()}</Text>
          
          {flowState === 'WAITING_ROOM' && roomId && (
            <View style={styles.roomInfo}>
              <Text style={styles.roomLabel}>Room ID:</Text>
              <Text style={styles.roomId}>{roomId.substring(0, 8)}...</Text>
            </View>
          )}
        </View>

        {/* Dots animation */}
        <View style={styles.dotsContainer}>
          <Animated.Text style={[styles.dot, { opacity: pulseValue }]}>●</Animated.Text>
          <Animated.Text style={[styles.dot, { opacity: pulseValue }]}>●</Animated.Text>
          <Animated.Text style={[styles.dot, { opacity: pulseValue }]}>●</Animated.Text>
        </View>

        {flowState === 'WAITING_ROOM' && (
          <Text style={styles.hint}>
            Finding you an opponent...{'\n'}
            This usually takes less than a minute
          </Text>
        )}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 60,
  },
  emoji: {
    fontSize: 64,
    marginBottom: 16,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#fff',
    letterSpacing: 2,
    marginBottom: 8,
  },
  username: {
    fontSize: 16,
    color: '#e94560',
    fontWeight: '600',
  },
  spinner: {
    width: 120,
    height: 120,
    backgroundColor: '#16213e',
    borderRadius: 60,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 40,
    borderWidth: 4,
    borderColor: '#0f3460',
  },
  spinnerText: {
    fontSize: 56,
  },
  statusContainer: {
    alignItems: 'center',
    marginBottom: 24,
  },
  statusText: {
    fontSize: 20,
    color: '#fff',
    fontWeight: '600',
    textAlign: 'center',
    marginBottom: 16,
  },
  roomInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#16213e',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
  },
  roomLabel: {
    fontSize: 14,
    color: '#aaa',
    marginRight: 8,
  },
  roomId: {
    fontSize: 14,
    color: '#e94560',
    fontWeight: 'bold',
  },
  dotsContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 8,
    marginBottom: 24,
  },
  dot: {
    fontSize: 20,
    color: '#e94560',
  },
  hint: {
    fontSize: 14,
    color: '#aaa',
    textAlign: 'center',
    lineHeight: 20,
  },
});