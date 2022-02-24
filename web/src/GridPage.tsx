import { Participant } from 'livekit-client';
import {
  AudioRenderer, LiveKitRoom, ParticipantView, StageProps,
} from 'livekit-react';
import React, { useEffect } from 'react';
import {
  onConnected, stopRecording, TemplateProps, useParams,
} from './common';
import styles from './GridPage.module.css';

export default function GridPage({ interfaceStyle }: TemplateProps) {
  const { url, token } = useParams();

  if (!url || !token) {
    return <div className="error">missing required params url and token</div>;
  }

  let containerClass = 'roomContainer';
  if (interfaceStyle) {
    containerClass += ` ${interfaceStyle}`;
  }
  return (
    <div className={containerClass}>
      <LiveKitRoom
        url={url}
        token={token}
        onConnected={onConnected}
        onLeave={stopRecording}
        stageRenderer={renderStage}
        connectOptions={{
          adaptiveStream: true,
        }}
      />
    </div>
  );
}

const renderStage: React.FC<StageProps> = ({ roomState }: StageProps) => {
  const {
    error, room, participants, audioTracks,
  } = roomState;
  const [visibleParticipants, setVisibleParticipants] = React.useState<Participant[]>([]);
  const [gridClass, setGridClass] = React.useState(styles.grid1x1);

  // select participants to display on first page, keeping ordering consistent if possible.
  useEffect(() => {
    let numVisible = participants.length;
    if (participants.length === 1) {
      setGridClass(styles.grid1x1);
    } else if (participants.length === 2) {
      setGridClass(styles.grid2x1);
    } else if (participants.length <= 4) {
      setGridClass(styles.grid2x2);
    } else if (participants.length <= 9) {
      setGridClass(styles.grid3x3);
    } else if (participants.length <= 16) {
      setGridClass(styles.grid4x4);
    } else {
      setGridClass(styles.grid5x5);
      numVisible = Math.min(numVisible, 25);
    }

    // remove any participants that are no longer connected
    const newParticipants: Participant[] = [];
    visibleParticipants.forEach((p) => {
      if (
        room?.participants.has(p.sid)
        || room?.localParticipant.sid === p.sid
      ) {
        newParticipants.push(p);
      }
    });

    // ensure active speakers are all visible
    room?.activeSpeakers?.forEach((speaker) => {
      if (
        newParticipants.includes(speaker)
        || (speaker !== room?.localParticipant
          && !room?.participants.has(speaker.sid))
      ) {
        return;
      }
      // find a non-active speaker and switch
      const idx = newParticipants.findIndex((p) => !p.isSpeaking);
      if (idx >= 0) {
        newParticipants[idx] = speaker;
      } else {
        newParticipants.push(speaker);
      }
    });

    // add other non speakers
    for (const p of participants) {
      if (newParticipants.length >= numVisible) {
        break;
      }
      if (newParticipants.includes(p) || p.isSpeaking) {
        continue;
      }
      newParticipants.push(p);
    }

    if (newParticipants.length > numVisible) {
      newParticipants.splice(numVisible, newParticipants.length - numVisible);
    }
    setVisibleParticipants(newParticipants);
  }, [participants]);

  if (error) {
    return <div className="error">{error}</div>;
  }

  if (!room) {
    return <div />;
  }

  if (visibleParticipants.length === 0) {
    return <div />;
  }

  const audioRenderers = audioTracks.map((track) => (
    <AudioRenderer key={track.sid} track={track} isLocal={false} />
  ));

  return (
    <div className={`${styles.stage} ${gridClass}`}>
      {visibleParticipants.map((participant) => (
        <ParticipantView
          key={participant.identity}
          participant={participant}
          orientation="landscape"
          width="100%"
          height="100%"
        />
      ))}
      {audioRenderers}
    </div>
  );
};
