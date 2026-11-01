def extend(schemas, op, paths, obj, ref, array, S, I, B, DATE, STAMP):
    schemas['ReportPeriod'] = obj(dict(from_=DATE, to=DATE, timezone=S, as_of=STAMP), ['from', 'to', 'timezone', 'as_of'])
    schemas['ReportPeriod']['properties']['from'] = schemas['ReportPeriod']['properties'].pop('from_')
    schemas['ReportRate'] = obj(dict(
        numerator=I,
        denominator=I,
        rate_percent={"type": "number", "format": "float", "nullable": True},
    ), ['numerator', 'denominator', 'rate_percent'])
    schemas['ReportSeatMinutes'] = obj(dict(
        booked=I,
        capacity=I,
        rate_percent={"type": "number", "format": "float", "nullable": True},
    ), ['booked', 'capacity', 'rate_percent'])
    schemas['OperationsReport'] = obj(dict(
        from_=DATE, to=DATE, timezone=S, as_of=STAMP,
        total_bookings=I,
        status_counts={"type": "object", "additionalProperties": I},
        member_attendance_count=I,
        pt_attendance_count=I,
        no_show=ref('ReportRate'),
        pending_completion_count=I,
        utilization=ref('ReportSeatMinutes'),
        cancellation_source_counts={"type": "object", "additionalProperties": I},
        unclassified_booking_count=I,
    ))
    schemas['OperationsReport']['properties']['from'] = schemas['OperationsReport']['properties'].pop('from_')
    schemas['TrainerReport'] = obj(dict(
        from_=DATE, to=DATE, timezone=S, as_of=STAMP, pt_id=I,
        rating_average={"type": "number", "format": "float", "nullable": True},
        review_count=I,
        attendance=ref('ReportRate'),
        excluded_count=I,
        unclassified_count=I,
        pending_completion_count=I,
    ))
    schemas['TrainerReport']['properties']['from'] = schemas['TrainerReport']['properties'].pop('from_')

    op('GET', '/reports/operations', 'Read facility operations metrics', roles='ADMIN or MANAGER', output='OperationsReport',
       filters=('from', 'to', 'room_id'),
       description='Inclusive Vietnam-local session_date range, default current month through today, at most 366 dates. Booking totals are current status at as_of and may include future sessions. Attendance and no-show metrics include ended sessions only. no_show is NO_SHOW / (COMPLETED + NO_SHOW), null for zero denominator. Utilization is booked seat-minutes for CONFIRMED/COMPLETED/NO_SHOW on ended OPEN sessions divided by capacity-minutes for those sessions; one booking consumes one seat regardless of PT attendance. It is reserved time, not measured workout time. Legacy ended bookings without confirmation/cancellation evidence are surfaced as unclassified. Invalid room ID returns 404; invalid date/filter returns 422.')
    op('GET', '/reports/pts/{id}', 'Read scoped PT performance metrics', roles='PT SELF or ADMIN/MANAGER ALL', output='TrainerReport',
       filters=('from', 'to'),
       description='Inclusive Vietnam-local session_date range, default current month through today, at most 366 dates. Rating aggregates valid reviews on completed ended bookings in the selected range. Attendance uses ended confirmed COMPLETED/NO_SHOW bookings and confirmed MEMBER/PT cancellations per BR-22; FACILITY cancellations, REJECTED, EXPIRED and pending requests are excluded. Missing confirmation or UNKNOWN cancellation evidence is unclassified. Pending end-of-session CONFIRMED bookings are reported separately. A zero denominator or no reviews returns null rather than zero. PT can read only their own profile ID; managers/admins may read any existing PT. Other roles receive 403; out-of-scope or missing profile receives 404.')
