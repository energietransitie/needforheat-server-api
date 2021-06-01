from datetime import timedelta
from enum import Enum
from typing import ClassVar, List, Optional

from pydantic import BaseModel, condecimal, conint, constr

from field import Datetime
from model import Account, Building


class HttpStatus(BaseModel):
    code: ClassVar[int]
    detail: str


class BadRequest(HttpStatus):
    code = 400


class Unauthorized(HttpStatus):
    code = 401


class Forbidden(HttpStatus):
    code = 403


class NotFound(HttpStatus):
    code = 404


class AccountLocation(BaseModel):
    longitude: condecimal(
        max_digits=Building.LOCATION_MAX_DIGITS,
        decimal_places=Building.LOCATION_DECIMAL_PLACES
    )
    latitude: condecimal(
        max_digits=Building.LOCATION_MAX_DIGITS,
        decimal_places=Building.LOCATION_DECIMAL_PLACES
    )


class AccountCreate(BaseModel):
    pseudonym: Optional[conint(ge=Account.PSEUDONYM_MIN, le=Account.PSEUDONYM_MAX)] = None
    location: Optional[AccountLocation] = None


class AccountItem(BaseModel):
    id: int
    pseudonym: int
    activation_token: str
    firebase_url: str


class AccountActivate(BaseModel):
    activation_token: str


class SessionToken(BaseModel):
    session_token: str


class AccountSession(SessionToken):
    pass


class DeviceSession(SessionToken):
    pass


class DeviceCreate(BaseModel):
    device_type: str
    device_activation_token: constr(strip_whitespace=True, min_length=8, max_length=1024)


class DeviceVerify(BaseModel):
    device_activation_token: constr(strip_whitespace=True, min_length=8, max_length=1024)


class DeviceTypeItem(BaseModel):
    name: str
    installation_manual_url: str

    class Config:
        orm_mode = True


class PropertyCompleteItem(BaseModel):
    id: int
    name: str
    unit: Optional[str]

    class Config:
        orm_mode = True


class DeviceTypeCompleteItem(BaseModel):
    id: int
    name: str
    installation_manual_url: str
    properties: List[PropertyCompleteItem]

    class Config:
        orm_mode = True


class DeviceItem(BaseModel):
    id: int
    device_type: DeviceTypeItem
    created_on: Datetime
    activated_on: Optional[Datetime]

    class Config:
        orm_mode = True


class DeviceItemMeasurementTime(DeviceItem):
    latest_measurement_timestamp: Optional[Datetime]

    class Config:
        orm_mode = True


class DeviceCompleteItem(BaseModel):
    id: int
    device_type: DeviceTypeCompleteItem
    device_activation_token: str
    created_on: Datetime
    activated_on: Optional[Datetime]

    class Config:
        orm_mode = True


class MeasurementValue(BaseModel):
    timestamp: Datetime
    value: str


class TimestampType(str, Enum):
    start = 'start'
    end = 'end'


class PropertyMeasurementsFixed(BaseModel):
    property_name: str
    timestamp: Datetime
    timestamp_type: TimestampType
    interval: timedelta
    measurements: List[str]


class PropertyMeasurementsVariable(BaseModel):
    property_name: str
    measurements: List[MeasurementValue]


class MeasurementsUploadFixed(BaseModel):
    upload_time: Datetime
    property_measurements: List[PropertyMeasurementsFixed]


class MeasurementsUploadVariable(BaseModel):
    upload_time: Datetime
    property_measurements: List[PropertyMeasurementsVariable]


class MeasurementsUploadResult(BaseModel):
    server_time: Datetime
    size: int
