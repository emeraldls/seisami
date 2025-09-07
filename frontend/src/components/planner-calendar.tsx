import { useState } from "react";
import { Calendar } from "~/components/ui/calendar";
import { Card, CardContent, CardFooter } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";

export function PlannerCalendar() {
  const [date, setDate] = useState<Date | undefined>(new Date());

  return (
    <Card className="w-full py-4 border rounded-none shadow-none">
      <CardContent className="px-4">
        <Calendar
          mode="single"
          selected={date}
          onSelect={setDate}
          className="bg-transparent p-0 [--cell-size:--spacing(10.5)]"
        />
      </CardContent>
      <CardFooter className="flex gap-2 border-t px-4 !pt-4 *:[div]:w-full">
        <Label htmlFor="time-from" className="Time whitespace-nowrap">
          Start Time
        </Label>
        <Input
          id="time-from"
          type="time"
          step="1"
          defaultValue="10:30:00"
          className="appearance-none w-fit [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
        />
      </CardFooter>
    </Card>
  );
}
