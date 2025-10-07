"use client"
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useUserStore } from '@/store/userStore'
import { useParticipantStore } from '@/store/participantStore'
import { useRoomStore } from '@/store/roomStore'
import type { Participant, Role,Room } from '@/types/models'
import { createRoom, joinRoom } from '@/api/rooms'
import { useRouter } from 'next/navigation'


export default function Dashboard() {
  const [isCreating, setIsCreating] = useState(false)
  const [isJoining, setIsJoining] = useState(false)
  const [isJoiningViaLink, setIsJoiningViaLink] = useState(false)
  const [roomName, setRoomName] = useState<string>("")
  const [roomLink, setRoomLink] = useState<string>("")
  const [error, setError] = useState<string>("")
  const setRoom = useRoomStore((state)=>state.setRoom)
  const room = useRoomStore((state)=>state.currentRoom)
  const setParticipant = useParticipantStore((state)=> state.setParticipant)
  const participant = useParticipantStore((state)=> state.currentParticipant)

  
  const user = useUserStore((state) => state.user)
  const router = useRouter()

  const clearError = () => setError("")

  async function handleCreateRoom() {
    if (!user || !roomName.trim()) {
      setError("User or Room Name is required")
      return
    }
    
    setIsCreating(true)
    setError("")
     
    try {
      const createData = await createRoom({
        userId: user.id, 
        username: user.name, 
        name: roomName.trim()
      })

      const room:Room = {
        roomId: createData.roomId,
        name:createData.name,
        createdAt: createData.createdAt,
        createdBy: createData.createdBy,
        hostURL: createData.hostURL,
        guestURL: createData.guestURL,
        audienceURL: createData.audienceURL,
      }

      setRoom(room)
      
    } catch (err) {
      console.error("Error creating/joining room:", err)
      setError("Failed to create room. Please try again.")
    } finally {
      setIsCreating(false)
    }
  }

  async function handleJoinRoom() {
    if (!user || !room || !room.hostURL) {
      setError("Incomplete room session data")
      return
    }
    
    setIsJoining(true)
    setError("")
    
    try {

      const {roomId,role} = extractRoomDetailsFromLink(room.hostURL)
      if (!roomId || !role) return


      const joinData = await joinRoom({
        userId: user.id, 
        roomId: roomId, 
        role: role as Role
      })
      
      const participant:Participant = {
        id:user.id,
        name:user.name,
        role:joinData.role,
        status:joinData.status,
        wsURL:joinData.wsURL
      }
      setParticipant(participant)

      if ( room && participant && participant.wsURL) {
        router.push(`/room/${room.roomId}`)
      }

    } catch (err) {
      console.error("Error joining room:", err)
      setError("Failed to join room. Please try again.")
    } finally {
      setIsJoining(false)
    }
  }

  async function handleJoinViaLink() {
    if (!user || !roomLink.trim()) {
      setError("Please enter a room link")
      return
    }
    
    setIsJoiningViaLink(true)
    setError("")
    
    try {
      const { roomId, role } = extractRoomDetailsFromLink(roomLink.trim())
      
      if (!roomId) {
        setError("Invalid room link format")
        return
      }
      
      const joinData = await joinRoom({
        userId: user.id, 
        roomId: roomId, 
        role: role as Role
      })
      
      const room:Room = {
        roomId: joinData.roomId,
        name: joinData.name,
        createdAt: joinData.createdAt,
        createdBy: joinData.createdBy,
        hostURL: null,
        guestURL: null,
        audienceURL: null,
      }

      setRoom(room)

      const participant:Participant = {
        id:user.id,
        name:user.name,
        role:joinData.role,
        status:joinData.status,
        wsURL:joinData.wsURL
      }
      setParticipant(participant)

      if ( room && participant && participant.wsURL) {
        router.push(`/room/${room.roomId}`)
      }
    } catch (err) {
      console.error("Error joining via link:", err)
      setError("Failed to join room via link. Please check the link and try again.")
    } finally {
      setIsJoiningViaLink(false)
    }
  }

  function extractRoomDetailsFromLink(link: string) {
    try {
      const url = new URL(link)
      const pathSegments = url.pathname.split('/')
      const joinIndex = pathSegments.indexOf('join')
      
      if (joinIndex === -1 || joinIndex + 1 >= pathSegments.length) {
        throw new Error('Invalid link format: missing join path or room ID')
      }
      
      const roomId = pathSegments[joinIndex + 1]
      const role = url.searchParams.get('role') || 'guest'
      
      return { roomId, role }
    } catch (error) {
      console.error('Error parsing room link:', error)
      return { roomId: null, role: null }
    }
  }

  if (!user) {
    return <div>Loading...</div>
  }

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-8">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-700">
          {error}
          <button 
            onClick={clearError}
            className="ml-2 text-red-500 hover:text-red-700"
          >
            ×
          </button>
        </div>
      )}
      
      <div className="space-y-4">
        <h2 className="text-lg font-semibold">Create New Room</h2>
        <div className="space-y-2">
          <Label htmlFor="roomName">Room Name</Label>
          <Input 
            id="roomName" 
            type="text" 
            value={roomName}
            onChange={(e) => setRoomName(e.target.value)}
            placeholder="Enter room name here..."
            disabled={isCreating}
          />
        </div>
        <Button 
          onClick={handleCreateRoom} 
          disabled={!roomName.trim() || isCreating}
        >
          {isCreating ? "Creating..." : "Create & Join Room"}
        </Button>
      </div>

      {room?.roomId && !participant?.wsURL && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Join Current Room</h2>
          <p className="text-sm text-gray-600">
            Room: {room.name} (ID: {room.roomId})
          </p>
          <Button 
            onClick={handleJoinRoom}
            disabled={isJoining}
          >
            {isJoining ? "Joining..." : "Join Room"}
          </Button>
        </div>
      )}

      <div className="space-y-4">
        <h2 className="text-lg font-semibold">Join via Link</h2>
        <div className="space-y-2">
          <Label htmlFor="link">Room Link</Label>
          <Input 
            id="link" 
            type="text" 
            value={roomLink}
            onChange={(e) => setRoomLink(e.target.value)}
            placeholder="Enter room link here..."
            disabled={isJoiningViaLink}
          />
        </div>
        <Button 
          onClick={handleJoinViaLink} 
          disabled={!roomLink.trim() || isJoiningViaLink}
        >
          {isJoiningViaLink ? "Joining..." : "Join"}
        </Button>
      </div>
    </div>
  )
}
