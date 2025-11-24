import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { COLORS, RADIUS, SHADOWS, SPACING, FONT_SIZES } from '../styles/colors';

interface ScoreBoardProps {
  scores: Record<string, number>;
  currentPlayer: string;
}

export default function ScoreBoard({ scores, currentPlayer }: ScoreBoardProps) {
  const players = Object.keys(scores);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>SCORE</Text>
      
      <View style={styles.scoresContainer}>
        {players.map((player) => {
          const isCurrentPlayer = player === currentPlayer;
          
          return (
            <View
              key={player}
              style={[
                styles.playerCard,
                isCurrentPlayer && styles.currentPlayerCard,
                !isCurrentPlayer && SHADOWS.small,
              ]}
            >
              <View style={styles.playerInfo}>
                <Text style={[
                  styles.playerIcon,
                  isCurrentPlayer && styles.currentPlayerText
                ]}>
                  {isCurrentPlayer ? '👤' : '🤖'}
                </Text>
                <Text style={[
                  styles.playerName,
                  isCurrentPlayer && styles.currentPlayerText
                ]}>
                  {player}
                </Text>
              </View>
              
              <Text style={[
                styles.score,
                isCurrentPlayer && styles.currentPlayerText
              ]}>
                {scores[player]}
              </Text>
            </View>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: COLORS.white,
    padding: SPACING.lg,
    borderRadius: RADIUS.large,
    marginBottom: SPACING.lg,
    ...SHADOWS.small,
  },
  title: {
    fontSize: FONT_SIZES.lg,
    fontWeight: 'bold',
    color: COLORS.textDark,
    textAlign: 'center',
    marginBottom: SPACING.md,
    letterSpacing: 2,
  },
  scoresContainer: {
    flexDirection: 'row',
    gap: SPACING.md,
  },
  playerCard: {
    flex: 1,
    backgroundColor: COLORS.background,
    padding: SPACING.md,
    borderRadius: RADIUS.medium,
    borderWidth: 3,
    borderColor: 'transparent',
  },
  currentPlayerCard: {
    backgroundColor: COLORS.primary,
    borderColor: COLORS.primaryDark,
    ...SHADOWS.medium,
  },
  playerInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: SPACING.sm,
    gap: SPACING.xs,
  },
  playerIcon: {
    fontSize: FONT_SIZES.lg,
  },
  playerName: {
    fontSize: FONT_SIZES.md,
    fontWeight: '600',
    color: COLORS.textDark,
    flex: 1,
  },
  currentPlayerText: {
    color: COLORS.white,
  },
  score: {
    fontSize: FONT_SIZES.xxxl,
    fontWeight: 'bold',
    textAlign: 'center',
    color: COLORS.primary,
  },
});