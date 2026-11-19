"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ProtectedRoute } from "@/components/protected-route";
import { useAuth } from "@/contexts/auth-context";
import { User, Mail, Phone, Calendar, Settings } from "lucide-react";
import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function ProfilePage() {
  return (
    <ProtectedRoute>
      <ProfileContent />
    </ProtectedRoute>
  );
}

function ProfileContent() {
  const { user, updateProfile } = useAuth();
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [draft, setDraft] = useState({ name: "", phone: "" });

  const openEditor = () => {
    if (!user) return;
    setDraft({ name: user.name ?? "", phone: user.phone ?? "" });
    setError("");
    setEditing(true);
  };

  const saveProfile = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (saving) return;
    setSaving(true);
    setError("");
    try {
      await updateProfile({ name: draft.name.trim(), phone: draft.phone.trim() });
      setEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save your profile. Please try again.");
    } finally {
      setSaving(false);
    }
  };

  if (!user) return null;

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white">Profile</h1>
            <p className="text-slate-600 dark:text-slate-300">
              Manage your account settings and preferences
            </p>
          </div>
          <Button className="bg-teal-600 hover:bg-teal-700" onClick={openEditor}>
            <Settings className="mr-2 h-4 w-4" />
            Edit Profile
          </Button>
        </div>

        <Dialog open={editing} onOpenChange={open => { if (!saving) setEditing(open); }}>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle>Edit Profile</DialogTitle>
              <DialogDescription>Update your personal details.</DialogDescription>
            </DialogHeader>
            <form onSubmit={saveProfile} className="space-y-4">
              {error && <Alert variant="destructive" role="alert"><AlertDescription>{error}</AlertDescription></Alert>}
              <div className="space-y-2">
                <Label htmlFor="profile-name">Full Name</Label>
                <Input id="profile-name" name="name" autoComplete="name" value={draft.name} onChange={event => setDraft(previous => ({ ...previous, name: event.target.value }))} maxLength={100} required disabled={saving} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="profile-email">Email</Label>
                <Input id="profile-email" type="email" value={user.email} readOnly aria-describedby="profile-email-note" className="bg-muted" />
                <p id="profile-email-note" className="text-sm text-muted-foreground">Your email is linked to your sign-in account.</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="profile-phone">Phone Number</Label>
                <Input id="profile-phone" name="phone" type="tel" autoComplete="tel" value={draft.phone} onChange={event => setDraft(previous => ({ ...previous, phone: event.target.value }))} maxLength={40} placeholder="+212 600 000 000" disabled={saving} />
              </div>
              <div className="flex justify-end gap-2">
                <Button type="button" variant="outline" disabled={saving} onClick={() => setEditing(false)}>Cancel</Button>
                <Button type="submit" className="bg-teal-600 hover:bg-teal-700" disabled={saving || !draft.name.trim()}>{saving ? "Saving..." : "Save Changes"}</Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>

        {/* Profile Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center space-x-4">
              <div className="w-16 h-16 bg-teal-100 rounded-full flex items-center justify-center">
                <User className="h-8 w-8 text-teal-600" />
              </div>
              <div>
                <CardTitle className="text-2xl">{user.name}</CardTitle>
                <CardDescription>Member since {new Date().toLocaleDateString()}</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <Mail className="h-4 w-4 text-slate-500" />
                  <span className="text-sm font-medium">Email</span>
                </div>
                <p className="text-slate-900 dark:text-white">{user.email}</p>
              </div>
              
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <Phone className="h-4 w-4 text-slate-500" />
                  <span className="text-sm font-medium">Phone</span>
                </div>
                <p className="text-slate-900 dark:text-white">{user.phone}</p>
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h3 className="text-lg font-semibold">Account Status</h3>
              <div className="flex space-x-2">
                <Badge variant="secondary" className="bg-green-100 text-green-800">
                  Active
                </Badge>
                <Badge variant="secondary" className="bg-blue-100 text-blue-800">
                  Verified
                </Badge>
              </div>
            </div>

            <Separator />

            <div className="space-y-4">
              <h3 className="text-lg font-semibold">Quick Actions</h3>
              <div className="grid gap-3 md:grid-cols-3">
                <Button variant="outline" className="justify-start">
                  <Calendar className="mr-2 h-4 w-4" />
                  Order History
                </Button>
                <Button variant="outline" className="justify-start">
                  <Settings className="mr-2 h-4 w-4" />
                  Preferences
                </Button>
                <Button variant="outline" className="justify-start">
                  <User className="mr-2 h-4 w-4" />
                  Security
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
