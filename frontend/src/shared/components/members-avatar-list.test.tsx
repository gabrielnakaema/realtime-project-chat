// @vitest-environment jsdom

import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { MembersAvatarList } from './members-avatar-list';

afterEach(() => {
  cleanup();
});

describe('MembersAvatarList', () => {
  it('shows compact two-letter initials and the remaining member count', () => {
    render(
      <MembersAvatarList
        members={[
          { user_id: '1', name: 'Alex Kim' },
          { user_id: '2', name: 'Jordan Lee' },
          { user_id: '3', name: 'Morgan Smith' },
        ]}
        max={2}
        variant="compactMuted"
      />,
    );

    expect(screen.getByLabelText('3 project members')).toBeTruthy();
    expect(screen.getByText('AK').className).toContain('bg-muted');
    expect(screen.getByText('JL')).toBeTruthy();
    expect(screen.getByText('+1')).toBeTruthy();
  });

  it('preserves the original avatar treatment by default', () => {
    render(<MembersAvatarList members={[{ user_id: '1', name: 'Alex Kim' }]} />);

    expect(screen.getByText('A').className).toContain('bg-primary');
  });
});
